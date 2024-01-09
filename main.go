package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"os"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections (for demonstration purposes)
		return true
	},
}
var connections = make(map[int]*websocket.Conn)

// функция запускается как горутина
func process(conns map[int]*websocket.Conn, n int) {
	var clientNo int
	//buf := make([]byte, 256)
	// получаем доступ к текущему соединению
	conn := conns[n]
	// определим, что перед выходом из функции, мы закроем соединение
	fmt.Println("Accept cnn:", n)
	defer conn.Close()

	for {

		messageType, buf, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		// Распечатываем полученое сообщение
		message := string(buf)
		//fmt.Printf("message %s", message)
		clientNo = 1
		// парсинг полученного сообщения
		//_, err = fmt.Sscanf(message, "%d", &clientNo) // определи номер клиента
		//if err != nil {
		//	// обработка ошибки формата
		//	conn.WriteMessage(websocket.TextMessage, []byte("error format message\n"))
		//	continue
		//}
		pos := strings.Index(message, " ") // нашли позицию разделителя

		//if pos > 0 {
		outMessage := message[pos+1:] // отчистили сообщение от номера клиента

		connTo := conns[clientNo] // This line is not needed for the WebSocket version

		if connTo == nil {
			conn.WriteMessage(websocket.TextMessage, []byte("client is close"))
			continue
		}
		var outBuf []byte
		switch messageType {
		case websocket.TextMessage:
			fmt.Println("Received text message:", string(message))
			outBuf = []byte(fmt.Sprintf("%d->>%s\n", clientNo, outMessage))
		case websocket.BinaryMessage:
			outBuf = buf
			fmt.Println("Received binary message:", "This is likely a binary file.")
		default:
			fmt.Println("Received unknown message type.")
		}

		//outBuf := []byte(fmt.Sprintf("%d->>%s\n", clientNo, outMessage))
		connTo.WriteMessage(websocket.TextMessage, []byte("hi"))

		// Отправить новую строку обратно клиенту
		err = connTo.WriteMessage(messageType, outBuf)
		// анализируем на ошибку
		if err != nil {
			fmt.Println("Error writing message:", err.Error())
			break // выходим из цикла
		}
		//}
	}
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}

	i := len(connections)
	if i < 0 {
		i = 0
	}
	connections[i] = conn

	// создаем пул соединений

	// сохраняем соединение в пул
	//conns[i] = conn

	// запускаем функцию process(conn)   как горутину
	go process(connections, i)
}

func handleFileTransfer(srcConn, destConn *websocket.Conn, filePath string) error {
	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a buffer to read and write chunks of the file
	buffer := make([]byte, 1024)

	for {
		// Read a chunk from the file
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			// End of file reached
			break
		} else if err != nil {
			return err
		}

		// Send the chunk to the destination connection
		err = destConn.WriteMessage(websocket.BinaryMessage, buffer[:bytesRead])
		if err != nil {
			return err
		}

		// Optionally, you can add a delay to control the rate of file transfer
		// time.Sleep(time.Millisecond * 10)
	}

	return nil
}

func main() {

	http.HandleFunc("/ws", websocketHandler)

	port := 8081
	fmt.Printf("Server is listening on :%d...\n", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
