package internal

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListHandler(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		invoices  []domain.Invoice
		finderErr error
		wantBody  string
		wantCode  int
	}{
		{
			name:  "200 ok with invoices",
			query: "?company_id=1&due_date=1970-01-01",
			invoices: []domain.Invoice{
				{
					IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
					Amount:    10000,
					Fee:       400,
					FeeRate:   0.04,
					Tax:       40,
					TaxRate:   0.10,
					DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
					Status:    domain.Processing,
				},
				{
					IssueDate: time.Date(1970, 1, 2, 9, 0, 0, 0, time.UTC),
					Amount:    5000,
					Fee:       200,
					FeeRate:   0.04,
					Tax:       20,
					TaxRate:   0.10,
					DueDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					Status:    domain.Processing,
				},
			},
			wantBody: `{"invoices":[{"issue_date":"1970-01-01T09:00:00Z","ammount":10000,"fee":400,"fee_rate":0.04,"tax":40,"tax_rate":0.1,"total":0,"due_date":"2024-10-30T00:00:00Z","status":"processing"},{"issue_date":"1970-01-02T09:00:00Z","ammount":5000,"fee":200,"fee_rate":0.04,"tax":20,"tax_rate":0.1,"total":0,"due_date":"2024-12-01T00:00:00Z","status":"processing"}]}` + "\n",
			wantCode: http.StatusOK,
		},
		{
			name:     "200 ok with no invoices",
			query:    "?company_id=1&due_date=1970-01-01",
			invoices: []domain.Invoice{},
			wantBody: `{"invoices":[]}` + "\n",
			wantCode: http.StatusOK,
		},
		{
			name:     "400 bad request without company_id",
			query:    "?company_id=&due_date=1970-01-01",
			invoices: []domain.Invoice{},
			wantBody: `{"message":"'company_id' mustn't be empty"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "400 bad request with invalid due_date",
			query:    "?company_id=1&due_date=INVALID",
			invoices: []domain.Invoice{},
			wantBody: `{"message":"Can't convert duedate parameter to date"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:      "500 internal server error when finder fails",
			query:     "?company_id=1&due_date=1970-01-01",
			finderErr: errors.New("this is test"),
			wantBody:  `{"message":"Failed to find invoices"}`,
			wantCode:  http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := FinderFunc(func(s string, t time.Time) ([]domain.Invoice, error) {
				return tt.invoices, tt.finderErr
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://localhost"+tt.query, nil)
			f := ListHandler(finder, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			f(w, r)

			assert.Equal(t, tt.wantCode, w.Code)

			b, err := io.ReadAll(w.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(b))
		})
	}
}
