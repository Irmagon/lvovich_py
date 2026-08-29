package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"fioincline/internal/server"
)

func main() {
	configPath := findConfig()
	cfg := server.ReadConfig(configPath)
	logPath := filepath.Join(filepath.Dir(configPath), "server.log")

	h := server.NewServer(cfg, logPath)
	defer h.Close()

	addr := fmt.Sprintf("%s:%d", cfg.Address, cfg.Port)

	msg := fmt.Sprintf("SOAP server running at http://%s:%d/soap", cfg.Address, cfg.Port)
	fmt.Println(msg)
	h.Log().Log("", msg)

	srv := &http.Server{Addr: addr, Handler: h}

	// Флаш лога при Ctrl+C / завершении (иначе потеряются последние строки буфера).
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		<-ch
		h.Close()
		os.Exit(0)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.Close() // log.Fatal зовёт os.Exit и обходит defer
		log.Fatal(err)
	}
}

// findConfig ищет config.ini: рядом с бинарником или в текущем каталоге.
func findConfig() string {
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "config.ini")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return "config.ini"
}
