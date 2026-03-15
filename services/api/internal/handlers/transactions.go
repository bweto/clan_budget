package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	api "clan-budget/services/api/gen"
)

// Server implements the generated ServerInterface.
type Server struct {
	db *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{db: db}
}

// ── Unimplemented stubs (required by ServerInterface) ──────────────────────────

func (s *Server) ListGroups(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) GetMonthlyReport(w http.ResponseWriter, r *http.Request, params api.GetMonthlyReportParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) ListTransactions(w http.ResponseWriter, r *http.Request, params api.ListTransactionsParams) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (s *Server) QuickAddTransaction(w http.ResponseWriter, r *http.Request) {
	// Superseded by the standard CreateTransaction form flow.
	w.WriteHeader(http.StatusNotImplemented)
}

// ── CreateTransaction ──────────────────────────────────────────────────────────

func (s *Server) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var body api.TransactionCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Could not decode request body")
		return
	}

	if body.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_amount", "Amount must be greater than 0")
		return
	}
	if body.Description == nil || *body.Description == "" {
		writeError(w, http.StatusBadRequest, "missing_description", "Description is required")
		return
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}

	txDate := body.Date.Time
	if txDate.IsZero() {
		txDate = time.Now()
	}

	var id string
	err := s.db.QueryRowContext(r.Context(),
		`INSERT INTO transactions (type, amount, currency, description, date, status)
		 VALUES ($1, $2, $3, $4, $5, 'completed')
		 RETURNING id`,
		string(body.Type), body.Amount, currency, body.Description, txDate,
	).Scan(&id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", "Failed to save transaction: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// ── helpers ────────────────────────────────────────────────────────────────────

func writeError(w http.ResponseWriter, code int, errCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{
		"code":    errCode,
		"message": message,
	})
}
