package server

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Config — конфигурация сервера из config.ini (порт script/soap/server.ts).
type Config struct {
	Address    string
	Port       int
	Token      string
	AllowedIPs []string
	Swagger    bool
	Logging    bool
	LogMode    string
	FlushMs    int
	BufferKB   int
}

// ReadConfig читает config.ini так же, как оригинальный readConfig().
// Путь — обычно корень репозитория (config.ini лежит рядом с сервером).
func ReadConfig(path string) Config {
	cfg := Config{Address: "0.0.0.0", Port: 3000, Swagger: true, Logging: true, LogMode: "async", FlushMs: 50, BufferKB: 64}
	ini, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	section := ""
	for _, raw := range strings.Split(string(ini), "\n") {
		s := strings.TrimSpace(raw)
		if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
			section = s[1 : len(s)-1]
			continue
		}
		if strings.HasPrefix(s, ";") || s == "" {
			continue
		}
		m := regexp.MustCompile(`^\s*(\w+)\s*=\s*(.+)\s*$`).FindStringSubmatch(s)
		if m == nil {
			continue
		}
		key, val := m[1], m[2]
		switch section {
		case "server":
			switch key {
			case "address":
				cfg.Address = val
			case "port":
				cfg.Port, _ = strconv.Atoi(val)
			case "swagger":
				cfg.Swagger = val == "true"
			}
		case "auth":
			switch key {
			case "token":
				cfg.Token = val
			case "allowed_ips":
				for _, ip := range strings.Split(val, ",") {
					ip = strings.TrimSpace(ip)
					if ip != "" {
						cfg.AllowedIPs = append(cfg.AllowedIPs, ip)
					}
				}
			}
		case "logging":
			switch key {
			case "enabled":
				cfg.Logging = val == "true"
			case "mode":
				cfg.LogMode = val
			case "flush_ms":
				cfg.FlushMs, _ = strconv.Atoi(val)
			case "buffer_kb":
				cfg.BufferKB, _ = strconv.Atoi(val)
			}
		}
	}
	return cfg
}
