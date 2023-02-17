package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
)

var kafkaTopic = "mailsucker_urls"

func main() {
	var kafkaHost = os.Getenv("KAFKA_HOST")
	http.HandleFunc("/urlapi", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var urlList []string

		if err := json.NewDecoder(r.Body).Decode(&urlList); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaHost, kafkaTopic, 0)
		if err != nil {
			log.Printf("Error dialing Kafka leader: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		messages := make([]kafka.Message, len(urlList))
		for i, url := range urlList {
			messages[i] = kafka.Message{Key: []byte(url)}
		}

		if _, err := conn.WriteMessages(messages...); err != nil {
			log.Printf("Error writing messages to Kafka: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
