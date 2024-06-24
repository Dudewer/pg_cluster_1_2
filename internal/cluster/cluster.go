package cluster

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"net/http"
	"pg_cluster_agent/internal/config"
	"strconv"
	"time"
)

type Cluster struct {
	MasterHost  *sql.DB
	SlaveHost   *sql.DB
	ClusterHost *sql.DB
	ArbiterHost string
}

func Init(cfg *config.Config) *Cluster {
	cluster := &Cluster{}
	timeout, err := strconv.Atoi(cfg.TIMEOUT)
	if err != nil {
		log.Info().Msg("TIMEOUT contains not number. Now TIMEOUT is default value = 60 sec")
		timeout = 60
	}
	timer := 0
	for timer <= timeout {
		var err error
		if cfg.MASTER_HOST != "" && cfg.MASTER_PORT != "" && cfg.MASTER_DB_NAME != "" {
			cluster.MasterHost, err = dbConnect(cfg.MASTER_HOST, cfg.MASTER_PORT, cfg.POSTGRES_USER, cfg.POSTGRES_PASSWORD, cfg.MASTER_DB_NAME)
		} else {
			log.Debug().Msgf("Something was wrong. Fill the .env file. MASTER_HOST or MASTER_PORT or MASTER_DB_NAME is empty")
		}
		if err != nil {
			log.Debug().Msg("Cannot connect to master.")
			time.Sleep(10 * time.Second)
			timer += 10
			continue
		}

		if cfg.SLAVE_HOST != "" && cfg.SLAVE_PORT != "" && cfg.SLAVE_DB_NAME != "" {
			cluster.SlaveHost, err = dbConnect(cfg.SLAVE_HOST, cfg.SLAVE_PORT, cfg.POSTGRES_USER, cfg.POSTGRES_PASSWORD, cfg.SLAVE_DB_NAME)
		} else {
			log.Debug().Msgf("Something was wrong. Fill the .env file. SLAVE_HOST or SLAVE_PORT or SLAVE_DB_NAME is empty")
		}
		if err != nil {
			log.Debug().Msg("Cannot connect to slave.")
			time.Sleep(10 * time.Second)
			timer += 10
			continue
		}

		if cfg.ARBITER_HOST != "" {
			cluster.ArbiterHost = cfg.ARBITER_HOST
		}

		if cfg.CLUSTER_HOST != "" && cfg.MASTER_PORT != "" && cfg.MASTER_DB_NAME != "" {
			cluster.ClusterHost, err = dbConnect(cfg.CLUSTER_HOST, cfg.MASTER_PORT, cfg.POSTGRES_USER, cfg.POSTGRES_PASSWORD, cfg.MASTER_DB_NAME)
		} else {
			log.Debug().Msgf("Something was wrong. Fill the .env file. CLUSTER_HOST or MASTER_PORT or MASTER_DB_NAME is empty")
		}
		if err != nil {
			log.Debug().Msg("Cannot connect to cluster.")
			time.Sleep(10 * time.Second)
			timer += 10
			continue
		}

		if err == nil {
			log.Info().Msg("Success init cluster")
			break
		}

		log.Info().Msg("Waiting for other hosts")
		time.Sleep(5 * time.Second)
		timer += 5
	}
	return cluster
}

func dbConnect(host, port, user, password, dbName string) (*sql.DB, error) {
	psqlURL := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)
	db, err := sql.Open("postgres", psqlURL)
	if err != nil {
		log.Error().Msgf("Can`t connect to database with host=%s port=%s user=%s dbname=%s and get: %s",
			host, port, user, dbName, err)
		return nil, err
	}
	log.Info().Msgf("Trying to ping database with host=%s port=%s user=%s dbname=%s",
		host, port, user, dbName)
	err = db.Ping()
	if err != nil {
		log.Error().Msgf("Can`t ping to database with host=%s port=%s user=%s dbname=%s and get: %s",
			host, port, user, dbName, err)
		return nil, err
	}
	return db, nil
}

func (ct *Cluster) PingMaster() bool {
	err := ct.MasterHost.Ping()
	if err != nil {
		log.Error().Msgf("Check Master. Error is %s", err)
	}
	log.Info().Bool("result", err == nil).Msg("Check Master")
	return err == nil
}

func (ct *Cluster) PingSlave() bool {
	err := ct.SlaveHost.Ping()
	if err != nil {
		log.Error().Msgf("Check Slave. Error is %s", err)
	}
	log.Info().Bool("result", err == nil).Msg("Check Slave")
	return err == nil
}

func (ct *Cluster) PingArbiter() (bool, error) {
	arbiterInfo := fmt.Sprintf("http://%s:8080/ping", ct.ArbiterHost)
	log.Info().Msgf("Now trying to do GET http://%s:8080/ping", ct.ArbiterHost)
	result, err := http.Get(arbiterInfo)
	if err != nil {
		log.Error().Msgf("Something wrong. We cant do HTTP Get. Error is %s", err)
		return false, err
	}
	if result.StatusCode != http.StatusOK {
		log.Error().Msgf("Status code of get to arbiter is %d.", result.StatusCode)
		return false, nil
	}
	log.Info().Bool("result", true).Msg("PingArbiter is OK")
	return true, nil
}

func (ct *Cluster) PingMasterFromArbiter() (bool, error) {
	arbiterInfo := fmt.Sprintf("http://%s:8080/master_status", ct.ArbiterHost)
	log.Info().Msgf("Now trying to do GET http://%s:8080/master_status", ct.ArbiterHost)
	result, err := http.Get(arbiterInfo)
	if err != nil {
		log.Error().Msgf("Something wrong. We cant do HTTP Get. Error is %s", err)
		return false, err
	}
	if result.StatusCode != http.StatusOK {
		log.Error().Msgf("Status code of get to arbiter is %d.", result.StatusCode)
		return false, nil
	}
	log.Info().Bool("return", true).Msg("MasterStatus is OK")
	return true, nil
}

func (ct *Cluster) Close() {
	ct.MasterHost.Close()
	ct.SlaveHost.Close()
	ct.ClusterHost.Close()
}
