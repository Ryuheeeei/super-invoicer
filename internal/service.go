package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
)

type Selector interface {
	Select(context.Context, string, time.Time) (*Rows, error)
}

type SelectorFunc func(context.Context, string, time.Time) (*Rows, error)

func (f SelectorFunc) Select(ctx context.Context, s string, t time.Time) (*Rows, error) {
	return f(ctx, s, t)
}

type FindService struct {
	Selector Selector
}

func (s *FindService) Find(ctx context.Context, companyID string, dueDate time.Time) ([]domain.Invoice, error) {
	rows, err := s.Selector.Select(ctx, companyID, dueDate)
	if err != nil {
		return nil, fmt.Errorf("find service error: %w", err)
	}
	invoices := make([]domain.Invoice, 0)
	for _, row := range rows.Rows {
		invoices = append(invoices,
			domain.Invoice{
				IssueDate: row.IssueDate,
				Amount:    row.Amount,
				Fee:       row.Fee,
				FeeRate:   row.FeeRate,
				Tax:       row.Tax,
				TaxRate:   row.TaxRate,
				DueDate:   row.DueDate,
				Status:    domain.Status(row.Status),
			},
		)
	}
	return invoices, nil
}
