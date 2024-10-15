package internal

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Ryuheeeei/super-invoicer/internal/domain"
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

func TestMySQL_Insert(t *testing.T) {
	tests := []struct {
		name    string
		row     []driver.Value
		wantErr error
	}{
		{
			name: "no error",
			row:  []driver.Value{"1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 10000, 400, float32(0.04), 40, float32(0.1), 10440, time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC), "processing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectBegin()
			mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")).ExpectExec().WithArgs(tt.row...).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			s := &MySQL{DB: db}
			_, err = s.Insert(context.Background(), "1", &domain.Invoice{
				IssueDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Amount:    10000,
				Fee:       400,
				FeeRate:   0.04,
				Tax:       40,
				TaxRate:   0.1,
				Total:     10440,
				DueDate:   time.Date(2024, 10, 31, 0, 0, 0, 0, time.UTC),
				Status:    domain.Processing,
			})
			assert.Equal(t, tt.wantErr, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
