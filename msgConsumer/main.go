package main

// now i gotta connect this to postgress and store all the messages in it

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type msg struct {
	MessageType string
	Data        string
	Reciever    string
	Sender      string
}

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
			fmt.Println("dindong comsumer : ", mes)
		}
	}()
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// connStr := "user=yash password=yash dbname=WebChats sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Initialize() error {
	if err := s.CreateMessageTable(); err != nil {
		return err
	}
	if err := s.CreateUserTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) CreateMessageTable() error {
	// messageId
	// messageData string
	// senderId
	// recieverId
	// sent dateTime
	// read dateTime

	query := `
	create table if not exists "Message"(
		message_id serial primary key,
		message_data text,
		sender_id integer,
		reciever_id integer,
		sentAt timestamp,
		readAt timestamp
	)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) CreateUserTable() error {
	// userId
	// username
	// userbio
	// profile_photo
	// bg_photo
	query := `
	create table if not exists "ChatUser"(
		user_id integer unique,
		username varchar(225),
		userbio varchar(225),
		profile_photo_link varchar(225)
	)`
	_, err := s.db.Exec(query)
	return err
}

var Rstore *RedisStore
var Pstore *PostgresStore

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	Pstore, err := NewPostgresStore()

	if err != nil {
		log.Fatal(err, " :: error while connecting postgress db")
	}
	err = Pstore.Initialize()
	if err != nil {
		log.Fatal(err, " :: error while initilizing postgress db")
	}
	Rstore, err = NewRedisStore()
	if err != nil {
		log.Fatal(err, "error while initilizing redis db db")
	}
	Rstore.SubRedis()
	wg.Wait()
}
