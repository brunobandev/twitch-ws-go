package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive: ", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	socketUrl := os.Getenv("TWITCH_WS_URL")
	socketUser := os.Getenv("TWITCH_YOUR_USERNAME")
	socketPass := os.Getenv("TWITCH_OAUTH_PASS")

	done = make(chan interface{})
	interrupt = make(chan os.Signal)

	signal.Notify(interrupt, os.Interrupt)

	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connection to websocket server: ", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("PASS %s", socketPass)))
	_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("NICK %s", socketUser)))
	_ = conn.WriteMessage(websocket.TextMessage, []byte("JOIN #eogabe"))

	for {

		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 1000):
			// Send an echo packet every second
			fmt.Println("Listening")
			// err := conn.WriteMessage(websocket.TextMessage, []byte("Sending message...!"))
			// if err != nil {
			// 	log.Println("Error during writing to websocket:", err)
			// 	return
			// }

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}
