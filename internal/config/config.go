package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	MCU      MCUConfig      `yaml:"mcu"`
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	StaticDir string `yaml:"static_dir"` // 前端静态文件目录，默认 "web/dist"
}

// DatabaseConfig MySQL 配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// MCUConfig 单片机通信配置
type MCUConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	HeartbeatSec     int    `yaml:"heartbeat_sec"`     // 心跳间隔（秒）
	HeartbeatTimeout int    `yaml:"heartbeat_timeout"` // 心跳超时（秒）
}

// DSN 返回 MySQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.DBName)
}

// Addr 返回 HTTP 监听地址
func (s *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// MCUAddr 返回单片机 TCP 监听地址
func (m *MCUConfig) Addr() string {
	return fmt.Sprintf("%s:%d", m.Host, m.Port)
}

// Load 从 YAML 和 .env 加载配置，优先级：.env > config.yaml > 默认值
func Load(yamlPath string) (*Config, error) {
	// 1. 默认值
	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Host:   "127.0.0.1",
			Port:   3306,
			User:   "root",
			DBName: "smart_cabinet",
		},
		MCU: MCUConfig{
			Host:             "0.0.0.0",
			Port:             9090,
			HeartbeatSec:     5,
			HeartbeatTimeout: 15,
		},
	}

	// 2. 加载 config.yaml
	data, err := os.ReadFile(yamlPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("读取 %s 失败: %w", yamlPath, err)
	}
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析 %s 失败: %w", yamlPath, err)
		}
	}

	// 3. 加载 .env（优先级最高）
	envPath := ".env"
	if yamlPath != "config.yaml" {
		// 如果指定了自定义 yaml，在同目录找 .env
		dir := ""
		if idx := strings.LastIndex(yamlPath, "/"); idx >= 0 {
			dir = yamlPath[:idx+1]
		}
		envPath = dir + ".env"
	}
	loadEnvFile(envPath)

	// 4. .env 覆盖对应的配置项
	applyEnvOverrides(cfg)

	return cfg, nil
}

// loadEnvFile 解析 .env 文件并设置环境变量
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env 不存在则跳过
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 KEY=VALUE
		idx := strings.Index(line, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// 去掉引号
		value = strings.Trim(value, `"'`)
		// 只在环境变量未设置时才设置（命令行优先级更高）
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// applyEnvOverrides 用环境变量覆盖配置
func applyEnvOverrides(cfg *Config) {
	lookup := map[string]func(string){
		// Server
		"SERVER_HOST": func(v string) { cfg.Server.Host = v },
		"SERVER_PORT": func(v string) { cfg.Server.Port = atoiOr(v, cfg.Server.Port) },

		// Database
		"DB_HOST":     func(v string) { cfg.Database.Host = v },
		"DB_PORT":     func(v string) { cfg.Database.Port = atoiOr(v, cfg.Database.Port) },
		"DB_USER":     func(v string) { cfg.Database.User = v },
		"DB_PASSWORD": func(v string) { cfg.Database.Password = v },
		"DB_NAME":     func(v string) { cfg.Database.DBName = v },

		// MCU
		"MCU_HOST": func(v string) { cfg.MCU.Host = v },
		"MCU_PORT": func(v string) { cfg.MCU.Port = atoiOr(v, cfg.MCU.Port) },
	}

	for key, apply := range lookup {
		if val, ok := os.LookupEnv(key); ok && val != "" {
			apply(val)
		}
	}
}

func atoiOr(s string, defaultVal int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
}
