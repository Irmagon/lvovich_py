package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func readLog(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}

func TestLoggerAsyncFlushes(t *testing.T) {
	dir := t.TempDir()
	l := NewLogger(filepath.Join(dir, "server.log"), "async", 20, 64)
	t.Cleanup(l.Close)

	l.Log("127.0.0.1", "GET /api/incline - {}")
	l.Log("", "SOAP Incline - null") // должен взять lastIP

	// Ждём флаша по таймеру.
	deadline := time.Now().Add(2 * time.Second)
	for {
		log := readLog(t, filepath.Join(dir, "server.log"))
		if strings.Count(log, "\n") >= 2 {
			t.Logf("log after flush:\n%s", log)
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("async flush не сработал за 2s; log=%q", log)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestLoggerAsyncCloseFlushesRemainder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	l := NewLogger(path, "async", 5000, 64) // большой интервал — флаш только по Close
	l.Log("127.0.0.1", "before close")
	l.Close()
	log := readLog(t, path)
	if !strings.Contains(log, "before close") {
		t.Fatalf("после Close() строка не сброшена: %q", log)
	}
}

func TestLoggerSyncWritesImmediately(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	l := NewLogger(path, "sync", 0, 0)
	t.Cleanup(l.Close)

	l.Log("127.0.0.1", "sync line")
	l.Close()

	log := readLog(t, path)
	if !strings.Contains(log, "sync line") {
		t.Fatalf("sync write не попал в файл: %q", log)
	}
	if !strings.Contains(log, "127.0.0.1") {
		t.Fatalf("IP не записан: %q", log)
	}
}

func TestLoggerLastIPFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	l := NewLogger(path, "sync", 0, 0)
	t.Cleanup(l.Close)

	l.Log("192.168.0.5", "first")
	l.Log("", "second") // no IP -> lastIP
	l.SetLastIP("10.0.0.1")
	l.Log("", "third")
	l.Close()

	log := readLog(t, path)
	if !strings.Contains(log, "192.168.0.5 first") {
		t.Fatalf("first: %q", log)
	}
	if !strings.Contains(log, "192.168.0.5 second") {
		t.Fatalf("second (lastIP должен подставить IP): %q", log)
	}
	if !strings.Contains(log, "10.0.0.1 third") {
		t.Fatalf("third (после SetLastIP): %q", log)
	}
}

func TestLoggerDisabledWritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	// enabled=false — логгер не должен создавать файл и писать в него.
	l := NewLogger(path, "async", 20, 64, false)
	l.Log("127.0.0.1", "GET /api/incline - {}")
	l.Log("", "SOAP Incline - null")
	l.Close()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("отключённый логгер не должен создавать файл: %v", err)
	}
	if got := readLog(t, path); got != "" {
		t.Fatalf("лог не должен быть пустым при отключённом логировании: %q", got)
	}
}

func TestLoggerBufferThresholdFlush(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "server.log")
	// buffer_kb=1 -> вызов Log() с большой строкой сразу превысит порог и сфлашит.
	l := NewLogger(path, "async", 5000, 1)
	t.Cleanup(l.Close)

	big := strings.Repeat("x", 2048)
	l.Log("127.0.0.1", big)

	// Флаш произошёл внепланово (по порогу), дожидаемся появления в файле.
	deadline := time.Now().Add(2 * time.Second)
	for {
		log := readLog(t, path)
		if strings.Contains(log, big) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("внеплановый флаш по порогу не сработал")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
