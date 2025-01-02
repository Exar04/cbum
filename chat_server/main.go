package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// upgrader is used to upgrade our http connection to ws conn
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	WebSocketToUsername = make(map[*websocket.Conn]string)
	UsernameToWebSocket = make(map[string]*websocket.Conn)
)

type RedisStore struct {
	db *redis.Client
}

func NewRedisStore() (*RedisStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:             fmt.Sprintf("%s:%s", "redis-ws", "6379"),
		Password:         "",
		DB:               0,
		DisableIndentity: true, // Disable set-info on connect
	})

	return &RedisStore{
		db: rdb,
	}, nil
}

func (s *RedisStore) PubRedis(ctx context.Context, mes msg) {
	messageJSON, err := json.Marshal(mes)
	if err != nil {
		log.Printf("Failed to serialize message: %v", err)
		return
	}

	s.db.Publish(ctx, "msg", messageJSON)
}

func (s *RedisStore) SubRedis() {
	subMsg := s.db.Subscribe(context.Background(), "msg")
	msgch := subMsg.Channel()

	go func() {
		for msgfromcha := range msgch {
			var mes msg
			err := json.Unmarshal([]byte(msgfromcha.Payload), &mes)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			if UsernameToWebSocket[mes.Reciever] != nil {
				UsernameToWebSocket[mes.Reciever].WriteJSON(mes)
			}
		}
	}()
}

var Rstore *RedisStore

func main() {
	Rstore, _ = NewRedisStore()
	Rstore.SubRedis()

	http.HandleFunc("/ws", SocketHandler)
	fmt.Println("web soc running on port 9000")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("there was a connection error : ", err)
		return
	}
	defer ws.Close()

	for {
		_, bytes, err := ws.ReadMessage()
		if err != nil {
			handleDisconnection(ws)
			break
		}
		err1 := handleIncomingMessage(ws, bytes)
		if err1 != nil {
			log.Print("Error handling message", err1)
		}
	}
	handleDisconnection(ws)
}

func handleDisconnection(sender *websocket.Conn) {
	user_id, _ := WebSocketToUsername[sender]
	delete(WebSocketToUsername, sender)
	delete(UsernameToWebSocket, user_id)
}

type msg struct {
	MessageType string
	Data        string
	Reciever    string
	Sender      string
}

const ChatMsg = "chatmsg"
const LoginMsg = "loginmsg"

func handleIncomingMessage(sender *websocket.Conn, data []byte) error {
	var DataRecieved msg
	err := json.Unmarshal(data, &DataRecieved)
	if err != nil {
		return err
	}
	fmt.Println(DataRecieved)

	switch DataRecieved.MessageType {
	case ChatMsg:
		Rstore.PubRedis(context.Background(), DataRecieved)

	case LoginMsg:
		if _, ok := UsernameToWebSocket[DataRecieved.Sender]; ok {
			sender.WriteJSON("User already exists")
			return nil
		}
		WebSocketToUsername[sender] = DataRecieved.Sender
		UsernameToWebSocket[DataRecieved.Sender] = sender
	}
	return nil
}

// {
//     "MessageType" : "loginmsg",
// 	"Data" :  "",
// 	"Reciever" : "",
// 	"Sender" : "tan"
// }

// {
//     "MessageType" : "chatmsg",
// 	"Data" :  "hiii",
// 	"Reciever" : "tan",
// 	"Sender" : "yan"
// }
