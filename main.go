package main

import (
	"log"

	"github.com/Misaka13906/FantasyGoWorld/internal/api/router"
	"github.com/Misaka13906/FantasyGoWorld/internal/config"
	"github.com/Misaka13906/FantasyGoWorld/pkg/logger"
)

func main() {
	logger.InitLogger()
	e := router.SetUpRouter()
	log.Println("Server Boot")
	e.Run(":" + config.Configs.HttpPort)
}
