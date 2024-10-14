package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
)

type InvoiceResponse struct {
	IssueDate time.Time `json:"issue_date"`
	Amount    int       `json:"ammount"`
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
