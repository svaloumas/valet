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
	URI           string
	conn          *amqp.Connection
	channel       *amqp.Channel
	queue         amqp.Queue
	delivery      <-chan amqp.Delivery
	publishParams config.PublishParams
	logger        *logrus.Logger
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
		cfg.QueueParams.Name,              // name
		cfg.QueueParams.Durable,           // durable
		cfg.QueueParams.DeletedWhenUnused, // delete when unused
		cfg.QueueParams.Exclusive,         // exclusive
		cfg.QueueParams.NoWait,            // no-wait
		nil,                               // arguments
	)
	if err != nil {
		panic(fmt.Sprintf("could not declare queue to RabbitMQ: %s", err))
	}
	delivery, err := channel.Consume(
		cfg.QueueParams.Name,        // queue
		cfg.ConsumeParams.Name,      // consumer
		cfg.ConsumeParams.AutoACK,   // auto-ack
		cfg.ConsumeParams.Exclusive, // exclusive
		cfg.ConsumeParams.NoLocal,   // no-local
		cfg.ConsumeParams.NoWait,    // no-wait
		nil,                         // args
	)
	if err != nil {
		panic(fmt.Sprintf("failed to init RabbitMQ queue consumer: %s", err))
	}
	rabbitmq := &rabbitmq{
		URI:           rabbitmqURI,
		conn:          conn,
		channel:       channel,
		queue:         queue,
		delivery:      delivery,
		publishParams: cfg.PublishParams,
		logger:        logger,
	}
	return rabbitmq
}

// Push adds a job to the queue.
func (q *rabbitmq) Push(j *domain.Job) error {
	body, err := json.Marshal(j)
	if err != nil {
		return err
	}
	err = q.channel.Publish(
		q.publishParams.Exchange,   // exchange
		q.publishParams.RoutingKey, // routing key
		q.publishParams.Mandatory,  // mandatory
		q.publishParams.Immediate,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	return nil
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
