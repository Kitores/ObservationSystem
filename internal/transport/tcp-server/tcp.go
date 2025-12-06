package tcp_server

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func StartTCPlogConsumer(network, address string) (conn net.Conn) {

	//conn, err := net.Dial(network, address)

	listener, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Println(err)
	}
	defer listener.Close()

	for {
		// Принятие входящих соединений
		conn, err = listener.Accept()
		if err != nil {
			log.Println("Ошибка принятия соединения:", err)
			continue
		}

		// Обработка соединения в отдельной горутине
		go handleConnections(conn)
	}
	return conn
}

func handleConnections(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	log.Printf("Клиент Подключился: %s\n", conn.RemoteAddr())
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Клиент отключился: %s\n", conn.RemoteAddr())
			return
		}

		log.Printf("Получено от %s: %s\n", conn.RemoteAddr(), message)

		response := fmt.Sprintf("%s", message)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return
		}
	}
}
