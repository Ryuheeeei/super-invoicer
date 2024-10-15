package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
			wantBody: `{"invoices":[{"issue_date":"1970-01-01T09:00:00Z","amount":10000,"fee":400,"fee_rate":0.04,"tax":40,"tax_rate":0.1,"total":0,"due_date":"2024-10-30T00:00:00Z","status":"processing"},{"issue_date":"1970-01-02T09:00:00Z","amount":5000,"fee":200,"fee_rate":0.04,"tax":20,"tax_rate":0.1,"total":0,"due_date":"2024-12-01T00:00:00Z","status":"processing"}]}` + "\n",
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
			finder := FinderFunc(func(context.Context, string, time.Time) ([]domain.Invoice, error) {
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

func TestCreateHandler(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		invoice       *domain.Invoice
		registererErr error
		wantBody      string
		wantCode      int
	}{
		{
			name: "200 ok with created invoice",
			body: `{"company_id":"1","amount":10000,"issue_date":"1970-01-01","due_date":"2024-10-30","status":"processing"}`,
			invoice: &domain.Invoice{
				IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
				Amount:    10000,
				Fee:       400,
				FeeRate:   0.04,
				Tax:       40,
				TaxRate:   0.10,
				DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
				Status:    domain.Processing,
			},
			wantBody: `{"issue_date":"1970-01-01T09:00:00Z","amount":10000,"fee":400,"fee_rate":0.04,"tax":40,"tax_rate":0.1,"total":0,"due_date":"2024-10-30T00:00:00Z","status":"processing"}` + "\n",
			wantCode: http.StatusOK,
		},
		{
			name:     "400 bad request when failed request body decode",
			body:     `INVALID`,
			wantBody: `{"message":"Failed to decode invoice request"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "400 bad request with empty company_id",
			body:     `{}`,
			wantBody: `{"message":"'company_id' mustn't be empty"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "400 bad request with invalid issue_date",
			body:     `{"company_id":"1","issue_date":"INVALID"}`,
			wantBody: `{"message":"Failed to decode issue_date as YYYY-MM-DD"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "400 bad request with invalid due_date",
			body:     `{"company_id":"1","issue_date":"1970-01-01","due_date":"INVALID"}`,
			wantBody: `{"message":"Failed to decode due_date as YYYY-MM-DD"}`,
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "400 bad request with invalid status",
			body:     `{"company_id":"1","issue_date":"1970-01-01","due_date":"1971-01-01","status":"UNKNOWN"}`,
			wantBody: fmt.Sprintf(`{"message":"'status' must be one of [unprocessed, processing, paid, error], but got %v"}`, "UNKNOWN"),
			wantCode: http.StatusBadRequest,
		},
		{
			name:          "500 internal server error when registerer fails",
			body:          `{"company_id":"1","amount":10000,"issue_date":"1970-01-01","due_date":"2024-10-30","status":"processing"}`,
			registererErr: errors.New("this is test"),
			wantBody:      `{"message":"Failed to create invoice"}`,
			wantCode:      http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registerer := RegistererFunc(func(context.Context, string, time.Time, int, time.Time, string) (*domain.Invoice, error) {
				return tt.invoice, tt.registererErr
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "http://localhost", strings.NewReader(tt.body))
			f := CreateHandler(registerer, slog.New(slog.NewTextHandler(os.Stderr, nil)))
			f(w, r)

			assert.Equal(t, tt.wantCode, w.Code)

			b, err := io.ReadAll(w.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(b))
		})
	}
}

func TestBasicAuthMiddleWware(t *testing.T) {
	username := "USERNAME"
	password := "PASSWORD"
	tests := []struct {
		name     string
		req      *http.Request
		handler  http.Handler
		wantCode int
		wantBody string
	}{
		{
			name: "next handler with valid credentials",
			req:  newRequestWithBasicAuth(username, password),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Next handler called"))
			}),
			wantCode: http.StatusOK,
			wantBody: "Next handler called",
		},
		{
			name: "401 unauthorized without credentials",
			req:  httptest.NewRequest(http.MethodGet, "http://localhost", nil),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.FailNow(t, "next handler should not be called")
			}),
			wantCode: http.StatusUnauthorized,
			wantBody: `{"message":"Authorization Header doesn't exist"}`,
		},
		{
			name: "401 unauthorized with invalid credentials",
			req:  newRequestWithBasicAuth(username, "INVALID"),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.FailNow(t, "next handler should not be called")
			}),
			wantCode: http.StatusUnauthorized,
			wantBody: `{"message":"Unauthorized"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			BasicAuthMiddleware(username, password, tt.handler).ServeHTTP(w, tt.req)

			assert.Equal(t, tt.wantCode, w.Code)

			b, err := io.ReadAll(w.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(b))
		})
	}
}

func newRequestWithBasicAuth(username, password string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	r.SetBasicAuth(username, password)
	return r
}
