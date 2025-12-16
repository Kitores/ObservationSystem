package app

import (
	"ObservationSystem/internal/logs"
	"ObservationSystem/internal/storage/postgre/methods"
	"ObservationSystem/internal/transport/rest/handlers/userHandlers"
	tcp_server "ObservationSystem/internal/transport/tcp-server"
	"fmt"
	"net/http"
	"sync"
)

func Run() {
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	//})
	connStr := "host=localhost port=5432 user=mihail password=secretPass123 dbname=logsdb sslmode=disable"
	storage, err := methods.NewPG(connStr)
	fmt.Println(storage)

	if err != nil {
		fmt.Printf("Error to connect postgres %w", err)
	}
	//Set logger

	logger := logs.LogInit()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		http.HandleFunc("/", userHandlers.UserHandler)

		logger.Info("HTTP Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("HTTP Server failed: %v", err)
		}
	}()

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
