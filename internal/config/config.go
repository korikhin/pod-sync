package config

import (
	"log"
	"strings"
	"time"

	"github.com/knadh/koanf"
	kenv "github.com/knadh/koanf/providers/env"
	kstr "github.com/knadh/koanf/providers/structs"
)

type Config struct {
	Sync    `koanf:"sync"`
	Storage `koanf:"storage"`
	HTTP    `koanf:"http"`
}

type Sync struct {
	Interval time.Duration `koanf:"interval"`
}

type Storage struct {
	URL            string        `koanf:"url"`
	MinConns       int           `koanf:"min-conns"`
	MaxConns       int           `koanf:"max-conns"`
	StartTimeout   time.Duration `koanf:"start-timeout"`
	ReadTimeout    time.Duration `koanf:"read-timeout"`
	WriteTimeout   time.Duration `koanf:"write-timeout"`
	IdleTimeout    time.Duration `koanf:"idle-timeout"`
	LifetimeJitter time.Duration `koanf:"lifetime-jitter"`
}

type HTTP struct {
	ReadTimeout     time.Duration `koanf:"read-timeout"`
	WriteTimeout    time.Duration `koanf:"write-timeout"`
	IdleTimeout     time.Duration `koanf:"idle-timeout"`
	ShutdownTimeout time.Duration `koanf:"shutdown-timeout"`
}

func defaultConfig() *Config {
	return &Config{
		Sync: Sync{
			Interval: 300 * time.Second,
		},
		Storage: Storage{
			MinConns:       1,
			MaxConns:       10,
			StartTimeout:   30 * time.Second,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			IdleTimeout:    30 * time.Minute,
			LifetimeJitter: 30 * time.Second,
		},
		HTTP: HTTP{
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    5 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
	}
}

const (
	envPrefix = "PSY__" // Pod SYnc
	delim     = "."
	Tag       = "koanf"
)

func envParser() func(string) string {
	return func(s string) string {
		s = strings.TrimPrefix(s, envPrefix)
		s = strings.Replace(s, "__", delim, -1)
		s = strings.Replace(s, "_", "-", -1)

		return strings.ToLower(s)
	}
}

func MustLoad() *Config {
	k := koanf.New(delim)
	cfg := defaultConfig()

	// Конфигурация по умолчанию
	if err := k.Load(kstr.Provider(cfg, Tag), nil); err != nil {
		log.Fatalf("error setting default config: %v", err)
	}
	// Конфигурация из переменных окружения
	if err := k.Load(kenv.Provider(envPrefix, delim, envParser()), nil); err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: Tag}); err != nil {
		log.Fatalf("error: %v", err)
	}

	return cfg
}
