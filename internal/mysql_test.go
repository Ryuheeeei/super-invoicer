package internal

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQL_Select(t *testing.T) {
	tests := []struct {
		name    string
		row     []driver.Value
		wantErr error
	}{
		{
			name: "no error",
			row:  []driver.Value{"1", "1", "2024-10-01", 10000, 400, 0.04, 40, 0.1, 0, "2024-10-31", "processing"},
		},
		{
			name:    "issue_date format error",
			row:     []driver.Value{"1", "1", "INVALID", 10000, 400, 0.04, 40, 0.1, 0, "2024-10-31", "processing"},
			wantErr: &time.ParseError{Layout: "2006-01-02", Value: "INVALID", LayoutElem: "2006", ValueElem: "INVALID", Message: ""},
		},
		{
			name:    "due_date format error",
			row:     []driver.Value{"1", "1", "2024-10-01", 10000, 400, 0.04, 40, 0.1, 0, "INVALID", "processing"},
			wantErr: &time.ParseError{Layout: "2006-01-02", Value: "INVALID", LayoutElem: "2006", ValueElem: "INVALID", Message: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta("SELECT invoice_id, company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status FROM invoice WHERE company_id = ? AND due_date BETWEEN ? AND ?;")).WithArgs("1", time.Now().Format(time.DateOnly), "9999-12-31").WillReturnRows(
				sqlmock.NewRows([]string{"invoice_id", "company_id", "issue_date", "amount", "fee", "fee_rate", "tax", "tax_rate", "total", "due_date", "status"}).AddRow(tt.row...))

			s := &MySQL{DB: db}
			_, err = s.Select(context.Background(), "1", time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC))
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
