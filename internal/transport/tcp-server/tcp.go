package tcp_server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	"log"
	"net"
)

func StartTCPlogConsumer(network, address string, storage *methods.PostgreSqlx) (conn net.Conn) {

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
		go handleConnections(conn, storage)
	}
	return conn
}

func handleConnections(conn net.Conn, pg *methods.PostgreSqlx) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	log.Printf("Клиент Подключился: %s\n", conn.RemoteAddr())

	var resp []byte
	conn.Read(resp) // Тут приходит данные для инициализации сервиса, а именно serviceName, host, team_owner, desc

	var newService entity.Service
	json.Unmarshal(resp, &newService) // DONE! -> И их нужно как-то распарсить из строки и передать в этот метод для записи нового сервиса в бд

	_, err := pg.RegisterService(newService)

	if err != nil {
		log.Println(err, "test")
	}
	var msg []byte
	for {
		_, err := reader.Read(msg)
		if err != nil {
			log.Printf("Клиент отключился: %s\n", conn.RemoteAddr())
			return
		}
		var logEntry entity.LogEntity
		err = json.Unmarshal(msg, &logEntry)
		if err != nil {
			log.Println(err, "test2")
			panic(err)
		}
		// Тут планируется парсинг из message лога, заполнения структуры LogEntity и сохранение новой записи в таблице

		pg.SaveLog(logEntry)

		//log.Printf("Получено от %s: %s\n", conn.RemoteAddr(), message)

		response := fmt.Sprintf("%s", msg)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return
		}
	}
}
