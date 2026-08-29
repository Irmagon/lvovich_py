package server

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// logLine — формат строки лога (как раньше): "[YYYY-MM-DD HH:MM:SS.mmm] ip msg".
type logLine struct {
	ts   string
	addr string
	msg  string
}

func (ln logLine) String() string {
	return fmt.Sprintf("[%s] %s %s\n", ln.ts, ln.addr, ln.msg)
}

// Logger пишет строки в server.log.
//
// В режиме async (по умолчанию) Log() только добавляет строку в in-memory буфер
// под коротким mutex, а фоновая горутина-флашер периодически сбрасывает буфер
// на диск пакетом. Это убирает диск из горячего пути запроса: параллельные
// запросы перестают упираться в мьютекс и файловый ввод-вывод.
// В режиме sync работает как раньше — запись на диск на каждый вызов.
type Logger struct {
	path    string
	mode    string
	flushMs time.Duration
	maxBuf  int
	enabled bool

	mu      sync.Mutex
	lastIP  string
	buf     []logLine
	bufSize int
	stopped bool

	stopCh  chan struct{}
	doneCh  chan struct{}
	started bool
}

// NewLogger создаёт логгер с указанным файлом (обычно server.log в корне репо).
// mode: "async" (по умолчанию) или "sync". flushMs — интервал флаша,
// bufferKB — порог объёма буфера (КБ) для внепланового флаша.
// enabled=false полностью отключает запись в файл.
func NewLogger(path, mode string, flushMs, bufferKB int, enabled ...bool) *Logger {
	if mode != "sync" {
		mode = "async"
	}
	if flushMs <= 0 {
		flushMs = 50
	}
	if bufferKB <= 0 {
		bufferKB = 64
	}
	on := true
	if len(enabled) > 0 {
		on = enabled[0]
	}
	l := &Logger{
		path:    path,
		mode:    mode,
		flushMs: time.Duration(flushMs) * time.Millisecond,
		maxBuf:  bufferKB * 1024,
		enabled: on,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	if mode == "async" && on {
		l.started = true
		go l.flusher()
	}
	return l
}

// SetLastIP запоминает IP последнего запроса (аналог _lastIp в оригинале).
func (l *Logger) SetLastIP(ip string) {
	l.mu.Lock()
	l.lastIP = ip
	l.mu.Unlock()
}

// Enabled сообщает, ведётся ли запись лога (enabled в конфиге).
// Позволяет вызывающей стороне полностью пропустить подготовку строки лога,
// когда логирование отключено.
func (l *Logger) Enabled() bool {
	return l.enabled
}

// Log пишет строку с IP последнего запроса (или '-').
// Если логгер отключён (enabled=false) — вызов не делает ничего.
func (l *Logger) Log(ip, msg string) {
	if !l.enabled {
		return
	}
	ln := logLine{
		ts:   nowStamp(),
		addr: "-",
		msg:  msg,
	}

	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		l.writeDirect(ln)
		return
	}
	if ip != "" {
		ln.addr = ip
	} else if l.lastIP != "" {
		ln.addr = l.lastIP
	}
	l.lastIP = ip

	if l.mode == "sync" {
		l.mu.Unlock()
		l.writeDirect(ln)
		return
	}

	l.buf = append(l.buf, ln)
	l.bufSize += len(ln.msg) + 40
	needFlush := l.bufSize >= l.maxBuf
	l.mu.Unlock()

	if needFlush {
		l.flush()
	}
}

// flush — синхронно сбрасывает накопленные строки на диск (для Close и порога).
func (l *Logger) flush() {
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return
	}
	batch := l.buf
	l.buf = nil
	l.bufSize = 0
	l.mu.Unlock()
	if len(batch) == 0 {
		return
	}
	l.writeBatch(batch)
}

// flusher — фоновая горутина: периодический сброс буфера на диск.
func (l *Logger) flusher() {
	t := time.NewTicker(l.flushMs)
	defer t.Stop()
	defer close(l.doneCh)
	for {
		select {
		case <-t.C:
			l.flush()
		case <-l.stopCh:
			l.flush()
			return
		}
	}
}

// Close останавливает флашер и сбрасывает остаток буфера на диск.
// Дальнейшие вызовы Log() пишут синхронно (best-effort).
func (l *Logger) Close() {
	if !l.enabled {
		return
	}
	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()
		return
	}
	l.stopped = true
	batch := l.buf
	l.buf = nil
	l.bufSize = 0
	started := l.started
	l.mu.Unlock()

	// Остановить флашер (если он запущен) ДО записи остатка, чтобы
	// flusher не конкурировал за файл.
	if started {
		close(l.stopCh)
		<-l.doneCh
	}
	if len(batch) > 0 {
		l.writeBatch(batch)
	}
}

// writeBatch открывает файл один раз и пишет партию строк.
func (l *Logger) writeBatch(batch []logLine) {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	var b []byte
	for _, ln := range batch {
		b = append(b, ln.String()...)
	}
	_, _ = f.Write(b)
}

// writeDirect — запись одной строки (sync-режим или после Close).
func (l *Logger) writeDirect(ln logLine) {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(ln.String())
}

// nowStamp — метка времени "YYYY-MM-DD HH:MM:SS.mmm".
func nowStamp() string {
	now := time.Now()
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1e6)
}
