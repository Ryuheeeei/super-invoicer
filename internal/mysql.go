package internal

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	DB *sql.DB
}

var _ Selector = (*MySQL)(nil)

type Rows struct {
	Rows []Row
}

type Row struct {
	InvoiceID string
	CompanyID string
	IssueDate time.Time
	Amount    int
	Fee       int
	FeeRate   float32
	Tax       int
	TaxRate   float32
	Total     int
	DueDate   time.Time
	Status    string
}

func (s *MySQL) Select(ctx context.Context, companyID string, dueDate time.Time) (*Rows, error) {
	var results []Row
	rows, err := s.DB.QueryContext(ctx, "SELECT invoice_id, company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status FROM invoice WHERE company_id = ? AND due_date BETWEEN ? AND ?;", companyID, time.Now().Format(time.DateOnly), dueDate.Format(time.DateOnly))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var row Row
		var issueDate string
		var dueDate string
		if err := rows.Scan(&row.InvoiceID, &row.CompanyID, &issueDate, &row.Amount, &row.Fee, &row.FeeRate, &row.Tax, &row.TaxRate, &row.Total, &dueDate, &row.Status); err != nil {
			break
		}
		row.IssueDate, err = time.ParseInLocation(time.DateOnly, issueDate, time.UTC)
		if err != nil {
			return nil, err
		}
		row.DueDate, err = time.ParseInLocation(time.DateOnly, dueDate, time.UTC)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return &Rows{Rows: results}, nil
}
