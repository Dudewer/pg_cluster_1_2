package main

import (
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"os/exec"
	arb "pg_cluster_agent/internal/arbiter"
	ct "pg_cluster_agent/internal/cluster"
	cfg "pg_cluster_agent/internal/config"
	mr "pg_cluster_agent/internal/master"
	sl "pg_cluster_agent/internal/slave"
	wr "pg_cluster_agent/internal/writer"
	"pg_cluster_agent/pkg/logger"
	"strings"
)

func main() {
	config, err := cfg.Load()
	if err != nil {
		log.Error().Msgf("Cant parse config. Error is %s", err)
		return
	}

	logger.Setup()
	log.Info().Msg("Success parsed config")

	if strings.ToLower(config.ROLE) == "master" {
		cmd := exec.Command("./add_ip.sh", config.CLUSTER_HOST)
		err := cmd.Run()
		if err != nil {
			log.Error().Msgf("Cant create cluster IP address. Error is %s", err)
		}
	}

	cluster := ct.Init(config)
	defer cluster.Close()

	switch strings.ToLower(config.ROLE) {
	case "arbiter":
		arb.RunArbiter(cluster)
	case "master":
		mr.RunMaster(cluster, config)
	case "slave":
		sl.RunSlave(cluster, config)
	case "writer":
		wr.RunWriter(cluster, config)
	}
}
