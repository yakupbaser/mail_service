package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/segmentio/kafka-go"
)

var kafkaHost = "localhost:9092"
var kafkaTopic = "mailsucker_html"
var kafkaMailTopic = "mails"

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if len(os.Args) > 1 {
		kafkaHost = os.Args[1]
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaHost},
		Topic:    kafkaTopic,
		GroupID:  "mailsucker-group",
		MaxBytes: 10e6,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaHost},
		Topic:    kafkaMailTopic,
		Balancer: &kafka.LeastBytes{},
	})

	for {
		select {
		case sig := <-signals:
			log.Printf("Stopping on signal: %v", sig)
			return
		default:
			m, err := reader.FetchMessage(context.Background())
			if err != nil {
				log.Printf("Failed to fetch message: %v", err)
				continue
			}

			html := string(m.Value)

			// Extract emails from the HTML using regex
			re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
			matches := re.FindAllString(html, -1)

			// Write each email to Kafka
			for _, match := range matches {
				err = writer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(m.Key),
					Value: []byte(match),
				})
				if err != nil {
					log.Printf("Failed to write message to Kafka: %v", err)
					return
				}

				log.Printf("Extracted mail from %s: %s", string(m.Key), match)
			}

			err = reader.CommitMessages(context.Background(), m)
			if err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}
