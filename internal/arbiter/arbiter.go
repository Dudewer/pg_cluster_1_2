package arbiter

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"net/http"
	"pg_cluster_agent/internal/cluster"
)

type Arbiter struct {
	ct *cluster.Cluster
}

func (a *Arbiter) MasterStatus(c *gin.Context) {
	result := a.ct.PingMaster()
	log.Info().Bool("res=", result).Msg("healthcheck master from arbiter")

	if !result {
		c.JSON(http.StatusBadGateway, gin.H{"Ping_Master": result})
	} else {
		c.JSON(http.StatusOK, gin.H{"Ping_Master": result})
	}
}

func (a *Arbiter) Ping(c *gin.Context) {
	log.Info().Str("client ip", c.ClientIP()).Msg("ping received ")
	c.JSON(http.StatusOK, "ping")
}

func RunArbiter(ct *cluster.Cluster) {
	log.Info().Msg("Run as Arbiter")
	handler := &Arbiter{ct: ct}

	server := gin.Default()
	server.GET("/master_status", handler.MasterStatus)
	server.GET("/ping", handler.Ping)

	server.Run(":8080")
}
