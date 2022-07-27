package main

import (
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"test-cluster/cluster"
	"test-cluster/config"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Print("join", config.Global.Cluster.NodeID, config.Global.Cluster.NodePort, config.Global.ServerPort)

	meta := cluster.NewMeta()
	cluster := cluster.NewCluster(meta)
	cluster.Join()

	log.Print("Starting web sever")
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/members", func(ctx *gin.Context) {
		ctx.IndentedJSON(http.StatusOK, GetMembers(cluster))
	})
	router.GET("/node/meta", func(ctx *gin.Context) {
		value := cluster.Memberlist.LocalNode().Meta
		ctx.IndentedJSON(http.StatusOK, gin.H{"meta": value})
	})
	router.PUT("/node/meta", func(ctx *gin.Context) {
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			log.Error().Err(err).Msg("receive node meta error")
		}
		defer ctx.Request.Body.Close()
		meta.SetNodeMeta(cluster.Memberlist.LocalNode().Name, body)
		cluster.Memberlist.UpdateNode(time.Second * 10)
		ctx.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.POST("/index", func(ctx *gin.Context) {
		for _, node := range cluster.Memberlist.Members() {
			err := cluster.Memberlist.SendReliable(node, []byte("create index"))
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		ctx.IndentedJSON(http.StatusOK, gin.H{"status": "ok"})
	})
	go router.Run("localhost:" + config.Global.ServerPort)
	log.Print("Started  web sever")

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := cluster.Memberlist.Leave(time.Second * 5); err != nil {
			panic(err)
		}
	}
}

func GetMembers(cluster *cluster.Cluster) map[string]string {
	var hostMap = make(map[string]string)
	for _, member := range cluster.Memberlist.Members() {
		hostMap[member.Name] = member.Address()
	}
	return hostMap
}
