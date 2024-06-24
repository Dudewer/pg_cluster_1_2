package slave

import (
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"os/exec"
	"pg_cluster_agent/internal/cluster"
	"pg_cluster_agent/internal/config"
	"time"
)

func RunSlave(ct *cluster.Cluster, cfg *config.Config) {
	log.Info().Msg("Run as Slave")
	for {
		time.Sleep(10 * time.Second)
		log.Info().Msg("Trying to ping master")
		checkMasterFromSlave := ct.PingMaster()
		if checkMasterFromSlave == true {
			log.Info().Msg("Ping master success")
		}
		log.Info().Msg("Trying to ping master from arbiter")
		checkMasterFromArbiter, err := ct.PingMasterFromArbiter()
		if err != nil {
			log.Error().Msgf("We cant do HTTP GET to Arbiter Host (%s). Error is %s",
				ct.ArbiterHost, err)
			continue
		}
		if checkMasterFromArbiter == true {
			log.Info().Msg("Ping master from arbiter success")
		}
		log.Info().Msg("Ping selfcheck")
		selfCheck := ct.SlaveHost.Ping()
		if selfCheck != nil {
			log.Debug().Msg("Self check is not OK")
		}
		if checkMasterFromArbiter == false && checkMasterFromSlave == false {
			log.Info().Msg("Promote to Master")

			cmd := exec.Command("./add_ip.sh", cfg.CLUSTER_HOST)
			err := cmd.Run()
			if err != nil {
				log.Error().Msg("Failed to create cluster ip")
			}
			cmd = exec.Command("gosu", "postgres", "pg_ctl", "promote", "-D", cfg.PGDATA)
			err = cmd.Run()

			if err == nil {
				log.Info().Msg("Success promote to Master")
				break
			}

			log.Error().Msgf("We could not promote this database to master. Error is %s", err)
		}
	}
}
