package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Starting maillogger")

	var postresHost string
	var brokers []string

	if isRunningInContainer() {
		postresHost = "postgres://user:password@postgres:5432/mydb?sslmode=disable"
		brokers = []string{os.Getenv("KAFKA_HOST")}
	} else {
		postresHost = "postgres://user:password@127.0.0.1:5432/mydb?sslmode=disable"
		brokers = []string{"localhost:29092"}
	}

	db, err := sql.Open("postgres", postresHost)

	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping Postgres: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mails (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	config := kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  "maillogger",
		Topic:    "mails",
		MinBytes: 10e3,
		MaxBytes: 10e6,
		MaxWait:  time.Second,
	}

	reader := kafka.NewReader(config)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signals:
			log.Println("Shutting down maillogger")
			return
		default:
			m, err := reader.FetchMessage(context.Background())
			if err != nil {
				log.Printf("Failed to fetch message: %v", err)
				continue
			}

			var email string = string(m.Value)

			// Check if the email already exists in the database
			row := db.QueryRow("SELECT email FROM mails WHERE email=$1", email)
			err = row.Scan(&email)

			if err != nil && err != sql.ErrNoRows {
				log.Printf("Failed to query email from Postgres: %v", err)
				reader.CommitMessages(context.Background(), m)
				break
			}

			if err == sql.ErrNoRows {
				_, err = db.Exec("INSERT INTO mails (email) VALUES ($1)", email)
				if err != nil {
					log.Printf("Failed to insert mail into Postgres: %v", err)
					reader.CommitMessages(context.Background(), m)
					break
				}

				log.Printf("%v mail address added to db", email)
			} else {
				log.Printf("%v already exists in database, skipping", email)
			}

			reader.CommitMessages(context.Background(), m)
		}
	}
}

func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return false
	}
	return true
}
