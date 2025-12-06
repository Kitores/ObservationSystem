package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:2000")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to server.")

	reader := bufio.NewReader(conn)
	go func() {
		for {
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Connection closed")
				return
			}
			fmt.Printf("Server: %s", response)
		}
	}()

	log.SetOutput(conn)

	// test
	ticker := time.NewTicker(5 * time.Second)
	for _ = range ticker.C {
		log.Println(".")
	}
}
