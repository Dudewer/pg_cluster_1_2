package writer

import (
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"pg_cluster_agent/internal/cluster"
	cfg "pg_cluster_agent/internal/config"
	"time"
)

var (
	Accept  = 0
	Dropped = 0
)

func WriterRUN(ct *cluster.Cluster, cfg *cfg.Config) {
	ipHost := cfg.CLUSTER_HOST
	time.Sleep(5 * time.Second)
	log.Info().Msg("Run as Writer")
	ct.MasterHost.Exec("CREATE TABLE IF NOT EXISTS test (id integer PRIMARY KEY);")
	acc, drop, potentialWroteNumbers, wroteNumbers := BenchTest(ct, ipHost)
	log.Info().Int("Accepted: ", acc).Msg("Results")
	log.Info().Int("Dropped: ", drop).Msg("Results")
	log.Info().Msgf("potentialWroteNumbers: ", potentialWroteNumbers)
	log.Info().Msgf("wroteNumbers: ", wroteNumbers)
}

func BenchTest(ct *cluster.Cluster, ipHost string) (accept, dropped int, potentialWroteNumbers []int, wroteNumbers []int) {
	log.Info().Msg("Run bench test")
	done := make(chan int)
	clusterHost := ct.ClusterHost
	numQueries := 1000000
	breakPoint := 700000
	//var potentialWroteNumbers []int
	//var wroteNumbers []int
	for i := 0; i < numQueries; i++ {
		go Write(clusterHost, i, done)
		potentialWroteNumbers = append(potentialWroteNumbers, i)
		if i == breakPoint {
			log.Info().Msg("Try to shutdown cluster host")
			ipClusterShutdown := fmt.Sprintf(ipHost, "/shutdown")
			resp, err := http.Get(ipClusterShutdown)
			if err != nil {
				log.Error().Msgf("We cant shutdown postgres of %s. Error is %s", ipClusterShutdown, err)
			}
			log.Info().Msgf("We get the %d status code.", resp.StatusCode)
		}
	}
	for i := 0; i < numQueries; i++ {
		val, ok := <-done
		if ok == false {
			log.Error().Msgf("Something was wrong.")
			break
		} else {
			if val == -1 {
				Dropped += 1
			} else {
				Accept += 1
				wroteNumbers = append(wroteNumbers, val)
			}
		}
	}
	return Accept, Dropped, potentialWroteNumbers, wroteNumbers
}

func Write(db *sql.DB, number int, done chan int) {
	log.Info().Int("number", number).Msg("num?")
	query := fmt.Sprintf("INSERT INTO public.test (id) VALUES (%d)", number)
	_, err := db.Exec(query)
	if err != nil {
		done <- -1
		log.Error().Msgf("Error database. Error is %s", err)
	}
	done <- number
}
