package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
)

type InvoiceResponse struct {
	IssueDate time.Time `json:"issue_date"`
	Amount    int       `json:"amount"`
	Fee       int       `json:"fee"`
	FeeRate   float32   `json:"fee_rate"`
	Tax       int       `json:"tax"`
	TaxRate   float32   `json:"tax_rate"`
	Total     int       `json:"total"`
	DueDate   time.Time `json:"due_date"`
	Status    string    `json:"status"`
}

type ListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

type Finder interface {
	Find(context.Context, string, time.Time) ([]domain.Invoice, error)
}

type FinderFunc func(context.Context, string, time.Time) ([]domain.Invoice, error)

func (f FinderFunc) Find(ctx context.Context, s string, date time.Time) ([]domain.Invoice, error) {
	return f(ctx, s, date)
}

func ListHandler(finder Finder, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		companyID := r.URL.Query().Get("company_id")
		if companyID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"'company_id' mustn't be empty"}`))
			return
		}
		dueDate, err := time.Parse(time.DateOnly, r.URL.Query().Get("due_date"))
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to convert duedate parameter to date", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Can't convert duedate parameter to date"}`))
			return
		}
		invoices, err := finder.Find(r.Context(), companyID, dueDate)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to find invoices", "customer_id", companyID, "due_date", dueDate, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Failed to find invoices"}`))
			return
		}

		resp := make([]InvoiceResponse, 0)
		for _, invoice := range invoices {
			resp = append(resp, InvoiceResponse{
				IssueDate: invoice.IssueDate,
				Amount:    invoice.Amount,
				Fee:       invoice.Fee,
				FeeRate:   invoice.FeeRate,
				Tax:       invoice.Tax,
				TaxRate:   invoice.TaxRate,
				Total:     invoice.Total,
				DueDate:   invoice.DueDate,
				Status:    string(invoice.Status),
			})
		}
		if err := json.NewEncoder(w).Encode(ListResponse{Invoices: resp}); err != nil {
			logger.ErrorContext(r.Context(), "Failed to encode found invoices to json", "invoices", invoices)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Failed to encode found invoices"}`))
			return
		}
	}
}

type InvoiceRequest struct {
	CompanyID string `json:"company_id"`
	IssueDate string `json:"issue_date"`
	Amount    int    `json:"amount"`
	DueDate   string `json:"due_date"`
	Status    string `json:"status"`
}

type Registerer interface {
	Register(context.Context, string, time.Time, int, time.Time, string) (*domain.Invoice, error)
}

type RegistererFunc func(context.Context, string, time.Time, int, time.Time, string) (*domain.Invoice, error)

func (f RegistererFunc) Register(ctx context.Context, s1 string, t1 time.Time, i int, t2 time.Time, s2 string) (*domain.Invoice, error) {
	return f(ctx, s1, t1, i, t2, s2)
}

func CreateHandler(registerer Registerer, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body InvoiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			logger.ErrorContext(r.Context(), "Failed to decode invoice request", "body", body, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Failed to decode invoice request"}`))
			return
		}
		if body.CompanyID == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"'company_id' mustn't be empty"}`))
			return
		}
		issueDate, err := time.ParseInLocation(time.DateOnly, body.IssueDate, time.UTC)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to decode issue_date as YYYY-MM-DD", "issue_date", body.IssueDate, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Failed to decode issue_date as YYYY-MM-DD"}`))
			return
		}
		dueDate, err := time.ParseInLocation(time.DateOnly, body.DueDate, time.UTC)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to decode due_date as YYYY-MM-DD", "due_date", body.DueDate, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message":"Failed to decode due_date as YYYY-MM-DD"}`))
			return
		}
		if body.Status != "unprocessed" && body.Status != "processing" && body.Status != "paid" && body.Status != "error" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"message":"'status' must be one of [unprocessed, processing, paid, error], but got %v"}`, body.Status)))
			return
		}
		invoice, err := registerer.Register(r.Context(), body.CompanyID, issueDate, body.Amount, dueDate, body.Status)
		if err != nil {
			logger.ErrorContext(r.Context(), "Failed to create invoice", "customer_id", body.CompanyID, "issue_date", body.IssueDate, "amount", body.Amount, "due_date", body.DueDate, "status", body.Status, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Failed to create invoice"}`))
			return
		}
		if err := json.NewEncoder(w).Encode(InvoiceResponse{IssueDate: invoice.IssueDate, Amount: invoice.Amount, Fee: invoice.Fee, FeeRate: invoice.FeeRate, Tax: invoice.Tax, TaxRate: invoice.TaxRate, Total: invoice.Total, DueDate: invoice.DueDate, Status: string(invoice.Status)}); err != nil {
			logger.ErrorContext(r.Context(), "Failed to encode created invoice to json", "invoice", invoice)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"Failed to encode created invoice"}`))
			return
		}
	}
}
