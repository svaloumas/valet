package jobqueue

import (
	"encoding/json"
	"fmt"
	"valet/internal/config"
	"valet/internal/core/domain"
	"valet/internal/core/port"
	"valet/pkg/env"
	"valet/pkg/log"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var _ port.JobQueue = &rabbitmq{}

type rabbitmq struct {
	URI      string
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    amqp.Queue
	delivery <-chan amqp.Delivery
	logger   *logrus.Logger
}

// NewRabbitMQ creates and returns a new fifoqueue instance.
func NewRabbitMQ(cfg config.RabbitMQ, loggingFormat string) *rabbitmq {
	logger := log.NewLogger("rabbitmq", loggingFormat)
	rabbitmqURI := env.LoadVar("RABBITMQ_URI")

	conn, err := amqp.Dial(rabbitmqURI)
	if err != nil {
		panic(fmt.Sprintf("could not connect to RabbitMQ: %s", err))
	}
	channel, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("could not open channel to RabbitMQ: %s", err))
	}
	queue, err := channel.QueueDeclare(
		cfg.QueueName,         // name
		cfg.Durable,           // durable
		cfg.DeletedWhenUnused, // delete when unused
		cfg.Exclusive,         // exclusive
		cfg.NoWait,            // no-wait
		nil,                   // arguments
	)
	if err != nil {
		panic(fmt.Sprintf("could not declare queue to RabbitMQ: %s", err))
	}
	delivery, err := channel.Consume(
		cfg.QueueName,       // queue
		"rabbitmq-consumer", // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		panic(fmt.Sprintf("failed to init RabbitMQ queue consumer: %s", err))
	}
	rabbitmq := &rabbitmq{
		URI:      rabbitmqURI,
		conn:     conn,
		channel:  channel,
		queue:    queue,
		delivery: delivery,
		logger:   logger,
	}
	return rabbitmq
}

// Push adds a job to the queue. Returns false if queue is full.
func (q *rabbitmq) Push(j *domain.Job) bool {
	body, err := json.Marshal(j)
	if err != nil {
		q.logger.Errorf("failed to push job to queue due to error while marshaling job: %s", err)
		return false
	}
	err = q.channel.Publish(
		"",           // exchange
		q.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		q.logger.Errorf("failed to push job to queue due to error while publishing message: %s", err)
		return false
	}
	return true
}

// Pop removes and returns the head job from the queue.
func (q *rabbitmq) Pop() *domain.Job {
	select {
	case msg := <-q.delivery:
		var j *domain.Job
		err := json.Unmarshal(msg.Body, &j)
		if err != nil {
			q.logger.Errorf("could not unmarshal message body: %s", err)
			return nil
		}
		return j
	default:
		return nil
	}
}

// Close liberates the bound resources of the job queue.
func (q *rabbitmq) Close() {
	q.channel.Close()
	q.conn.Close()
}
