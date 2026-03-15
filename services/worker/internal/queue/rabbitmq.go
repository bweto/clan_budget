package queue

import (
	"encoding/json"
	"log"

	"clan-budget/services/worker/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

const (
	ExchangeName    = "recurring_exchange"
	QueueName       = "recurring_queue"
	RoutingKey      = "recurring.run"
	
	DLXExchangeName = "dlx_exchange"
	DLXQueueName    = "dlq_queue"
	DLXRoutingKey   = "recurring.dlq"
)

type RecurringMessage struct {
	RuleID string  `json:"rule_id"`
	UserID string  `json:"user_id"`
	GroupID string `json:"group_id"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
	Desc   string  `json:"description"`
}

func Connect() *RabbitMQ {
	log.Println("Connecting to RabbitMQ...")
	conn, err := amqp091.Dial(config.GetRabbitMQConnStr())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	mq := &RabbitMQ{Conn: conn, Channel: ch}
	mq.setupTopology()

	return mq
}

func (mq *RabbitMQ) setupTopology() {
	// 1. Setup Dead Letter Exchange and Queue
	err := mq.Channel.ExchangeDeclare(DLXExchangeName, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLX: %v", err)
	}

	_, err = mq.Channel.QueueDeclare(DLXQueueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare DLQ: %v", err)
	}

	err = mq.Channel.QueueBind(DLXQueueName, DLXRoutingKey, DLXExchangeName, false, nil)
	if err != nil {
		log.Fatalf("Failed to bind DLQ: %v", err)
	}

	// 2. Setup Main Exchange and Queue with DLX configuration
	args := amqp091.Table{
		"x-dead-letter-exchange":    DLXExchangeName,
		"x-dead-letter-routing-key": DLXRoutingKey,
	}

	err = mq.Channel.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare Main Exchange: %v", err)
	}

	_, err = mq.Channel.QueueDeclare(QueueName, true, false, false, false, args)
	if err != nil {
		log.Fatalf("Failed to declare Main Queue: %v", err)
	}

	err = mq.Channel.QueueBind(QueueName, RoutingKey, ExchangeName, false, nil)
	if err != nil {
		log.Fatalf("Failed to bind Main Queue: %v", err)
	}
	
	log.Println("RabbitMQ Topology setup complete (with DLQ).")
}

func (mq *RabbitMQ) Publish(msg RecurringMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return mq.Channel.Publish(
		ExchangeName,
		RoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)
}

func (mq *RabbitMQ) Close() {
	if mq.Channel != nil {
		mq.Channel.Close()
	}
	if mq.Conn != nil {
		mq.Conn.Close()
	}
}
