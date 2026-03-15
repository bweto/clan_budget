package processor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"clan-budget/services/worker/internal/db"
	"clan-budget/services/worker/internal/queue"

	"github.com/cenkalti/backoff/v4"
	"github.com/rabbitmq/amqp091-go"
)

type Processor struct {
	mq          *queue.RabbitMQ
	db          *db.Database
	concurrency int
}

func NewProcessor(mq *queue.RabbitMQ, db *db.Database, concurrency int) *Processor {
	return &Processor{
		mq:          mq,
		db:          db,
		concurrency: concurrency,
	}
}

func (p *Processor) Start(ctx context.Context) {
	// Set QoS rules for backpressure
	err := p.mq.Channel.Qos(
		p.concurrency, // prefetch count matches concurrency
		0,             // prefetch size
		false,         // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	msgs, err := p.mq.Channel.Consume(
		queue.QueueName, // queue
		"",              // consumer
		false,           // auto-ack (IMPORTANT: Must be false for DLQ/Retries)
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf("Started RabbitMQ Consumer with concurrency %d", p.concurrency)

	for i := 0; i < p.concurrency; i++ {
		go p.worker(ctx, msgs, i)
	}
}

func (p *Processor) worker(ctx context.Context, msgs <-chan amqp091.Delivery, id int) {
	log.Printf("Worker %d started", id)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopped", id)
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			p.handleMessage(ctx, d, id)
		}
	}
}

func (p *Processor) handleMessage(ctx context.Context, d amqp091.Delivery, workerID int) {
	var msg queue.RecurringMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("[Worker %d] Failed to unmarshal message: %v. Sending to DLQ.", workerID, err)
		// Reject without requeue sends it to DLX
		d.Reject(false)
		return
	}

	log.Printf("[Worker %d] Processing Rule ID: %s", workerID, msg.RuleID)

	// Setup exponential backoff for Postgres Insertions
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.MaxElapsedTime = 10 * time.Second // Retry for roughly ~10 seconds before giving up

	operation := func() error {
		// Enforce pending_revision status logic via explicit DB query
		_, err := p.db.Conn.ExecContext(ctx,
			`INSERT INTO transactions (user_id, group_id, recurring_rule_id, type, amount, currency, description, date, status) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending_revision')`,
			msg.UserID, msg.GroupID, msg.RuleID, msg.Type, msg.Amount, "USD", msg.Desc, time.Now(),
		)
		return err
	}

	err := backoff.RetryNotify(operation, b, func(err error, t time.Duration) {
		log.Printf("[Worker %d] DB Insert failed for Rule %s: %v. Retrying in %v...", workerID, msg.RuleID, err, t)
	})

	if err != nil {
		log.Printf("[Worker %d] Fatal DB error for Rule %s after retries: %v. Sending to DLQ.", workerID, msg.RuleID, err)
		d.Reject(false) // Send to DLQ
		return
	}

	log.Printf("[Worker %d] Successfully generated pending transaction for Rule %s", workerID, msg.RuleID)
	d.Ack(false) // Success
}
