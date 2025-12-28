package app

import (
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/logs"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	http_server "github.com/Kitores/ObservationSystem/internal/transport/http-server"
	tcp_server "github.com/Kitores/ObservationSystem/internal/transport/tcp-server"
	"net/http"
	"sync"
)

func Run() {

	//Connection db
	connStr := "host=localhost port=5432 user=mihail password=secretPass123 dbname=logsdb sslmode=disable"
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
		http_server.StartHttpServer(storage)

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
