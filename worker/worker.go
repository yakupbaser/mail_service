package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/gocolly/colly"
	"github.com/segmentio/kafka-go"
)

var kafkaHost = "localhost:9092"
var kafkaTopic = "mailsucker_urls"
var kafkaMailTopic = "mails"

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	if len(os.Args) > 1 {
		kafkaHost = os.Args[1]
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaHost},
		Topic:   kafkaTopic,
		GroupID: "mailsucker-group",
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaHost},
		Topic:    kafkaMailTopic,
		Balancer: &kafka.LeastBytes{},
	})

	c := colly.NewCollector()

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if strings.Contains(link, "mailto:") {
			mail := strings.TrimPrefix(link, "mailto:")
			if isValidMail(mail) {
				err := writer.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte(e.Request.URL.String()),
					Value: []byte(mail),
				})
				if err != nil {
					log.Printf("Failed to write message to Kafka: %v", err)
					return
				}

				log.Printf("Extracted mail from %s: %s", e.Request.URL.String(), mail)
			}
		}
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

			url := string(m.Key)

			err = c.Visit(url)
			if err != nil {
				log.Printf("Failed to visit URL %s: %v", url, err)
				continue
			}

			err = reader.CommitMessages(context.Background(), m)
			if err != nil {
				log.Printf("Failed to commit message: %v", err)
			}
		}
	}
}
func isValidMail(mail string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(mail)
}
