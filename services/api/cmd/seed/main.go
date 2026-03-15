package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("Environment variable %s is not set", key)
	}
	return value
}

func getDBConnStr() string {
	user := getEnv("POSTGRES_USER")
	password := getEnv("POSTGRES_PASSWORD")
	dbname := getEnv("POSTGRES_DB")
	host := getEnv("POSTGRES_HOST")
	port := getEnv("POSTGRES_PORT")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
}

func main() {
	// Attempt to load .env file from common locations (ignore errors if file not found)
	paths := []string{".env", "../../.env", "../../../.env", "../../../../.env"}
	for _, p := range paths {
		_ = godotenv.Load(p)
	}

	// Initialize gofakeit
	gofakeit.Seed(time.Now().UnixNano())

	// Connect to the DB
	log.Println("Connecting to the database...")
	db, err := sql.Open("postgres", getDBConnStr())
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}

	ctx := context.Background()

	// Begin Transaction
	log.Println("Beginning transaction...")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Safety defer to rollback if panic or early return
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Panic occurred, rolling back: %v", p)
			tx.Rollback()
			os.Exit(1)
		} else if err != nil {
			log.Printf("Error occurred, rolling back: %v", err)
			tx.Rollback()
		} else {
			log.Println("Committing transaction...")
			err = tx.Commit()
			if err != nil {
				log.Fatalf("Failed to commit transaction: %v", err)
			}
			log.Println("Seed completed successfully!")
		}
	}()

	// 1. Create Users
	log.Println("Seeding 5 users...")
	var userIDs []string
	for i := 0; i < 5; i++ {
		email := gofakeit.Email()
		var id string
		err = tx.QueryRowContext(ctx, "INSERT INTO users (email) VALUES ($1) RETURNING id", email).Scan(&id)
		if err != nil {
			return // Return maps to err rollback
		}
		userIDs = append(userIDs, id)
	}

	// 2. Create Family Groups
	log.Println("Seeding 2 family groups...")
	var groupIDs []string
	for i := 0; i < 2; i++ {
		name := gofakeit.LastName() + " Family"
		creatorID := userIDs[i] // Give the first 2 users creation rights
		var id string
		err = tx.QueryRowContext(ctx, "INSERT INTO family_groups (name, created_by) VALUES ($1, $2) RETURNING id", name, creatorID).Scan(&id)
		if err != nil {
			return
		}
		groupIDs = append(groupIDs, id)
	}

	// 3. Create Group Members
	log.Println("Seeding group memberships...")
	// Group 0 -> Users 0, 1, 2
	memberships := []struct {
		groupID string
		userID  string
		role    string
	}{
		{groupIDs[0], userIDs[0], "admin"},
		{groupIDs[0], userIDs[1], "member"},
		{groupIDs[0], userIDs[2], "member"},
		// Group 1 -> Users 3, 4
		{groupIDs[1], userIDs[3], "admin"},
		{groupIDs[1], userIDs[4], "member"},
	}

	for _, m := range memberships {
		_, err = tx.ExecContext(ctx, "INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)", m.groupID, m.userID, m.role)
		if err != nil {
			return
		}
	}

	// 4. Create Recurring Rules
	log.Println("Seeding 3 recurring rules...")
	type Rule struct {
		ID      string
		GroupID string
		Type    string
		Amount  float64
		Name    string
	}
	var rules []Rule

	ruleDefs := []struct {
		name    string
		amount  float64
		typeStr string
		groupID string
		creator string
	}{
		{"Internet Bill", 60.00, "expense", groupIDs[0], userIDs[0]},
		{"Rent", 1200.00, "expense", groupIDs[0], userIDs[0]},
		{"Electricity", 80.00, "expense", groupIDs[1], userIDs[3]},
	}

	now := time.Now()
	for _, r := range ruleDefs {
		var id string
		err = tx.QueryRowContext(ctx,
			`INSERT INTO recurring_rules (group_id, created_by, type, amount, currency, description, frequency, start_date) 
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			r.groupID, r.creator, r.typeStr, r.amount, "USD", r.name, "monthly", now.AddDate(-1, 0, 0),
		).Scan(&id)
		if err != nil {
			return
		}
		rules = append(rules, Rule{ID: id, GroupID: r.groupID, Type: r.typeStr, Amount: r.amount, Name: r.name})
	}

	// 5. Generate past transactions based on recurring rules (last 3 months)
	log.Println("Generating recent transactions for recurring rules...")
	for _, r := range rules {
		for i := 0; i < 3; i++ {
			// +/- 5% variance for utilities, fixed for rent
			amount := r.Amount
			if r.Name != "Rent" {
				variance := amount * 0.05
				amount = amount + (gofakeit.Float64Range(-variance, variance))
			}
			date := now.AddDate(0, -i, 0)

			// Find an admin for this group to assign the transaction to
			var userID string
			for _, m := range memberships {
				if m.groupID == r.GroupID && m.role == "admin" {
					userID = m.userID
					break
				}
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO transactions (user_id, group_id, recurring_rule_id, type, amount, currency, description, date) 
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				userID, r.GroupID, r.ID, r.Type, math.Round(amount*100)/100, "USD", r.Name, date,
			)
			if err != nil {
				return
			}
		}
	}

	// 6. Generate 1000 standalone transactions over the past 365 days
	log.Println("Generating 1000 standalone transactions...")

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO transactions (user_id, group_id, type, amount, currency, description, date) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for i := 0; i < 1000; i++ {
		userID := userIDs[rand.Intn(len(userIDs))]

		// Find which group this user belongs to
		var groupID string
		for _, m := range memberships {
			if m.userID == userID {
				groupID = m.groupID
				break
			}
		}

		// 80% expenses, 20% incomes
		recordType := "expense"
		if rand.Float32() < 0.2 {
			recordType = "income"
		}

		amount := gofakeit.Float64Range(5.0, 500.0)
		description := gofakeit.ProductName()
		if recordType == "income" {
			amount = gofakeit.Float64Range(100.0, 3000.0)
			description = "Deposit / Payment"
		}

		date := gofakeit.DateRange(now.AddDate(-1, 0, 0), now)

		_, err = stmt.ExecContext(ctx, userID, groupID, recordType, math.Round(amount*100)/100, "USD", description, date)
		if err != nil {
			return
		}
	}
}
