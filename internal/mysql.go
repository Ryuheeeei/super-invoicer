package internal

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
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
	rows, err := s.DB.QueryContext(ctx, "SELECT invoice_id, company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status FROM invoice WHERE company_id = ? AND due_date BETWEEN ? AND ? AND status != 'paid';", companyID, time.Now().Format(time.DateOnly), dueDate.Format(time.DateOnly))
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

func (s *MySQL) Insert(ctx context.Context, companyID string, invoice *domain.Invoice) (*Row, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // The rollback will be ignored if the tx has been committed later in the function.

	stmt, err := s.DB.PrepareContext(ctx, "INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, companyID, invoice.IssueDate, invoice.Amount, invoice.Fee, invoice.FeeRate, invoice.Tax, invoice.TaxRate, invoice.Total, invoice.DueDate, invoice.Status)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// get auto-incremented invoice_id
	invoiceID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Row{
		InvoiceID: strconv.FormatInt(invoiceID, 10),
		CompanyID: companyID,
		IssueDate: invoice.IssueDate,
		Amount:    invoice.Amount,
		Fee:       invoice.Fee,
		FeeRate:   invoice.FeeRate,
		Tax:       invoice.Tax,
		TaxRate:   invoice.TaxRate,
		Total:     invoice.Total,
		DueDate:   invoice.DueDate,
		Status:    string(invoice.Status),
	}, nil
}
