package logs

import (
	"bufio"
	"fmt"
	"net"
)

type LogWriter struct {
}

func (l LogWriter) Write(p []byte) (n int, err error) {
	_, err = l.Write(p)
	if err != nil {
		fmt.Println("Error sending:", err)
	}
	return fmt.Print(string(p))
}

func ReadTcp(conn net.TCPConn) {
	buf := make([]byte, 1024)
	conn.Read(buf)
}

func StartTCPlogConsumer(network, address string) (conn net.Conn) {

	conn, err := net.Dial(network, address)

	listener, err := net.Listen("tcp", ":2000")
	if err != nil {
		fmt.Println(err)
	}
	defer listener.Close()
	fmt.Println("Сервер запущен на :2000")

	for {
		// Принятие входящих соединений
		conn, err = listener.Accept()
		if err != nil {
			fmt.Println("Ошибка принятия соединения:", err)
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

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Клиент отключился: %s\n", conn.RemoteAddr())
			return
		}

		fmt.Printf("Получено от %s: %s\n", conn.RemoteAddr(), message)

		response := fmt.Sprintf("%s", message)
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Ошибка отправки:", err)
			return
		}
	}
}
