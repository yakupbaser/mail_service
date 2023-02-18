package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
)

var kafkaTopic = "mailsucker_html"

func main() {
	log.Println("Starting mailsucker")

	var kafkaHost string

	if isRunningInContainer() {
		kafkaHost = os.Getenv("KAFKA_HOST")
	} else {
		kafkaHost = "localhost:9092"
	}

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

		for _, url := range urlList {
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Error getting URL: %v", err)
				continue
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body: %v", err)
				continue
			}

			if len(body) > 0 {
				chunkSize := 1024000 // 1 MB
				for i := 0; i < len(body); i += chunkSize {
					end := i + chunkSize
					if end > len(body) {
						end = len(body)
					}
					message := kafka.Message{
						Key:   []byte(url),
						Value: body[i:end],
					}
					if _, err := conn.WriteMessages(message); err != nil {
						log.Printf("Error writing message to Kafka: %v", err)
						continue
					}
				}
			}

			resp.Body.Close()
		}

		w.WriteHeader(http.StatusOK)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err != nil {
		return false
	}
	return true
}
