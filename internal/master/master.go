package master

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"os/exec"
	"pg_cluster_agent/internal/cluster"
	"pg_cluster_agent/internal/config"
)

func MasterRUN(ct *cluster.Cluster, cfg *config.Config) {
	log.Info().Msg("Run as Master")
	cmd := exec.Command("./add_ip.sh", cfg.CLUSTER_HOST)
	err := cmd.Run()
	if err != nil {
		log.Error().Msgf("Cant create IP address. Error is %s", err)
	}
	server := gin.Default()
	server.GET("/shutdown", Shutdown)
	server.GET("/accept", Accept)
}

func Accept(_ *gin.Context) {
	cmd := exec.Command("iptables", "-F")
	err := cmd.Run()
	if err == nil {
		log.Info().Msg("Success block connections to Master")
	}
}

func Shutdown(_ *gin.Context) {
	cfg, err := config.Load()
	if err != nil {
		log.Err(err)
	}
	err = exec.Command("./del_ip.sh", cfg.CLUSTER_HOST).Run()
	if err != nil {
		log.Err(err).Msg("Cant delete cluster ip")
	}
	err = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--dport", "5432", "-j", "DROP").Run()
	if err != nil {
		log.Err(err).Msg("Cannot block input d5432 connections to Master")
	}
	err = exec.Command("iptables", "-A", "INPUT", "-p", "tcp", "--sport", "5432", "-j", "DROP").Run()
	if err != nil {
		log.Err(err).Msg("Cannot block input s5432 connections to Master")
	}
}
