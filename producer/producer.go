package main

import (
	"encoding/json"
	"fmt"
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/IBM/sarama"
)



type Comment struct{
	Text string `form: "text" json:"text"`
}


func main() {
	app := fiber.New()
	api := app.Group("/api/v1")
	api.Post("/comments", createComment)
	app.Listen(":3000")
}


func ConnectProducer(brokerUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	conn, err := sarama.NewSyncProducer(brokerUrl, config)
	if err != nil {

		return nil, err
	}
	return conn, nil

}

func PushCommentToQueue(topic string, message []byte) error {
	brokerUrl := []string{"localhost:29092"}
	producer , err:= ConnectProducer(brokerUrl)

	if err != nil {
		return err
	}
	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),

	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err

	}

	fmt.Printf("Message is stored in the topic %s/partition %d/offset %d\n", topic, partition, offset)
	return nil
}


func createComment(c *fiber.Ctx) error {

	cmt := new(Comment)
	if err := c.BodyParser(cmt); err != nil {
		log.Println(err)
		c.Status(400).JSON(&fiber.Map{
			"success":false,
			"message": err,
		})
		return err
	}
	cmtByted, err := json.Marshal(cmt)
	PushCommentToQueue("comments", cmtByted)

	c.JSON(&fiber.Map{
		"success": true,
		"message": "All is ok",
		"comment": cmt,
	})
	if err != nil {
		c.Status(500).JSON(&fiber.Map{
			"success":false,
			"message": "Error while creating comment",
		})
		return err
	}

	return err
}