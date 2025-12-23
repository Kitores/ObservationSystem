package tcp_server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Kitores/ObservationSystem/internal/models/entity"
	"github.com/Kitores/ObservationSystem/internal/storage/postgre/methods"
	"io"
	"log"
	"net"
	"time"
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

	waitingNewService(conn, pg)

	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Ошибка чтения: %v\n", err)
			}
			log.Printf("Клиент отключился: %s\n", conn.RemoteAddr())
			return
		}
		// Пропускаем пустые сообщения
		if len(msg) == 0 || len(msg) == 1 && msg[0] == '\n' {
			continue
		}
		var logEntity entity.LogEntity
		json.Unmarshal(msg, &logEntity)
		logEntity.ReceivedAt = time.Now().UTC()
		fmt.Println(logEntity)

		// Тут планируется парсинг из message лога, заполнения структуры LogEntity и сохранение новой записи в таблице

		//pg.SaveLog(logEntry)

		log.Printf("Получено от %s: %s\n", conn.RemoteAddr(), msg)

		response := fmt.Sprintf("%s", msg)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return
		}
	}
}

func waitingNewService(conn net.Conn, pg *methods.PostgreSqlx) {
	reader := bufio.NewReader(conn)
	var serviceId int64 = 0
	var hostId int64 = 0
	for {

		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Ошибка чтения: %v\n", err)
			}
			log.Printf("Клиент отключился: %s\n", conn.RemoteAddr())
			return
		}
		// Пропускаем пустые сообщения
		if len(msg) == 0 || len(msg) == 1 && msg[0] == '\n' {
			continue
		}

		// Тут приходит данные для инициализации сервиса, а именно serviceName, host, team_owner, desc

		var newService entity.Service
		json.Unmarshal(msg, &newService) // DONE! -> И их нужно как-то распарсить из строки и передать в этот метод для записи нового сервиса в бд

		if newService.IsFirst == true {
			fmt.Println(newService, "REQ SERVICE")
			//serviceId, hostId, err = pg.RegisterService(newService)
			fmt.Println(serviceId, hostId)
			if err != nil {
				log.Println(err, "test")
			}
			fmt.Println("Работает мать его!")
			return
		}

		//time.Sleep(5 * time.Second)

		//if serviceId != 0 {
		//	break
		//}

	}

}
