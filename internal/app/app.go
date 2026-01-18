package app

import (
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/config"
	"github.com/Kitores/ObservationSystem/internal/logs"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	http_server "github.com/Kitores/ObservationSystem/internal/transport/http-server"
	tcp_server "github.com/Kitores/ObservationSystem/internal/transport/tcp-server"
	"net/http"
	"sync"
)

func Run() {

	//Getting Config
	cfg, err := config.InitConfig()

	//Connection db
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB, cfg.PostgresSSLMode)
	storage, err := methods.NewPG(connStr)

	if err != nil {
		fmt.Printf("Error to connect postgresql %w", err)
	}

	//Set custom logger
	logger := logs.LogInit()

	var wg sync.WaitGroup

	// HTTP-server start listening
	wg.Add(1)
	go func() {
		http_server.StartHttpServer(cfg, storage)

		logger.Info("HTTP Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("HTTP Server failed: %v", err)
		}
	}()

	// TCP-server start listening
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("TCP Log Consumer starting on :2000")
		conn := tcp_server.StartTCPlogConsumer("tcp", ":2000", storage)
		defer conn.Close()

	}()

	logger.Info("All servers starting...")
	wg.Wait()
}
