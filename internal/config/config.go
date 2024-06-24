package config

import (
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Config struct {
	ARBITER_HOST      string `koanf:"ARBITER_HOST"`
	ROLE              string `koanf:"ROLE"`
	CLUSTER_HOST      string `koanf:"CLUSTER_HOST"`
	POSTGRES_USER     string `koanf:"POSTGRES_USER"`
	POSTGRES_PASSWORD string `koanf:"POSTGRES_PASSWORD"`
	MASTER_HOST       string `koanf:"MASTER_HOST"`
	MASTER_PORT       string `koanf:"MASTER_PORT"`
	MASTER_DB_NAME    string `koanf:"MASTER_DB_NAME"`
	SLAVE_HOST        string `koanf:"SLAVE_HOST"`
	SLAVE_PORT        string `koanf:"SLAVE_PORT"`
	SLAVE_DB_NAME     string `koanf:"SLAVE_DB_NAME"`
	PGDATA            string `koanf:"PGDATA"`
	TIMEOUT           string `koanf:"TIMEOUT"`
}

func Load() (config *Config, err error) {
	k := koanf.New(".")

	err = k.Load(env.Provider("", ".", func(s string) string { return s }), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load environment variables")
	}
	if err := k.Unmarshal("", &config); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshall config")
	}

	return config, nil
}
