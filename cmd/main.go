package main

import (
	"ObservationSystem/internal/http-server/handlers/userHandlers"
	"ObservationSystem/internal/logs"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func main() {
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	//})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		http.HandleFunc("/", userHandlers.UserHandler)

		log.Println("HTTP Server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP Server failed: %v", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("TCP Log Consumer starting on :2000")
		conn := logs.StartTCPlogConsumer("tcp", ":2000")
		defer conn.Close()

		fmt.Println("TCP Log Consumer started successfully")

	}()

	log.Println("All servers starting...")
	wg.Wait()

}


// TODO: Попробовать подключить к notification system (но сначала запушить наверное чтоб можно было подключить пакет из удалённого репозитория) именно е Tcp серверу