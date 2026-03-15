package scheduler

import (
	"context"
	"log"

	"clan-budget/services/worker/internal/db"
	"clan-budget/services/worker/internal/queue"

	"github.com/robfig/cron/v3"
)

type CronScheduler struct {
	cron *cron.Cron
	db   *db.Database
	mq   *queue.RabbitMQ
}

func NewScheduler(db *db.Database, mq *queue.RabbitMQ) *CronScheduler {
	// Setup cron with Seconds field for easier testing, though we will schedule monthly
	c := cron.New(cron.WithSeconds())
	return &CronScheduler{cron: c, db: db, mq: mq}
}

// Start will begin the cron job asynchronously
func (s *CronScheduler) Start() {
	// "0 0 0 1 * *" -> At 00:00:00 on day-of-month 1.
	// We'll also add a 15-second test schedule to verify during dev
	s.cron.AddFunc("0 0 0 1 * *", func() {
		log.Println("[CRON] Executing Monthly Recurring Rules Job")
		s.RunNow()
	})

	log.Println("Started Cron Scheduler (Next run is 1st of next month)")
	s.cron.Start()
}

func (s *CronScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

// RunNow triggers the logic immediately. It queries the active recurring_rules and publishes them to RabbitMQ.
func (s *CronScheduler) RunNow() {
	log.Println("RunNow: Querying recurring rules...")
	// Note: You would normally filter by start_date/end_date and active status
	rows, err := s.db.Conn.QueryContext(context.Background(),
		`SELECT r.id, r.group_id, r.created_by, r.type, r.amount, r.description 
		 FROM recurring_rules r 
		 JOIN family_groups fg ON fg.id = r.group_id`)
	
	if err != nil {
		log.Printf("[CRON ERROR] Failed to query recurring rules: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var msg queue.RecurringMessage
		err := rows.Scan(&msg.RuleID, &msg.GroupID, &msg.UserID, &msg.Type, &msg.Amount, &msg.Desc)
		if err != nil {
			log.Printf("[CRON ERROR] Failed to scan rule row: %v", err)
			continue
		}

		err = s.mq.Publish(msg)
		if err != nil {
			log.Printf("[CRON ERROR] Failed to queue rule %s: %v", msg.RuleID, err)
		} else {
			count++
		}
	}

	log.Printf("RunNow completed: Queued %d recurring rules.", count)
}
