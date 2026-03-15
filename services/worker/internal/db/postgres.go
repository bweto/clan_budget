package db

import (
	"context"
	"database/sql"
	"log"
	"time"

	"clan-budget/services/worker/internal/config"

	_ "github.com/lib/pq"
)

type Database struct {
	Conn *sql.DB
}

func Connect() *Database {
	log.Println("Connecting to PostgreSQL...")
	db, err := sql.Open("postgres", config.GetDBConnStr())
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}
	
	log.Println("PostgreSQL connection established.")
	return &Database{Conn: db}
}

func (d *Database) Close() {
	if d.Conn != nil {
		d.Conn.Close()
	}
}
