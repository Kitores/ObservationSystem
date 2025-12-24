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

	serviceID, hostID := waitingNewService(conn, pg)

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
		logEntity.HostID = hostID
		logEntity.ServiceID = serviceID

		fmt.Println(logEntity)

		// Тут планируется парсинг из message лога, заполнения структуры LogEntity и сохранение новой записи в таблице

		saveLog, err := pg.SaveLog(logEntity)
		if err != nil {
			log.Println(err, saveLog)
		}

		log.Printf("Получено от %s: %s\n", conn.RemoteAddr(), msg)

		response := fmt.Sprintf("%s", msg)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return
		}
	}
}

func waitingNewService(conn net.Conn, pg *methods.PostgreSqlx) (serviceId, hostId int64) {
	reader := bufio.NewReader(conn)
	serviceId = 0
	hostId = 0
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
			newService.HostName = "windows11"
			serviceId, hostId, err = pg.RegisterService(newService)
			fmt.Println(serviceId, hostId)
			if err != nil {
				log.Println(err, "test")
			}
			fmt.Println("Работает мать его!")
			return serviceId, hostId
		}

		//time.Sleep(5 * time.Second)

		//if serviceId != 0 {
		//	break
		//}

	}
}
