package internal

import (
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
)

type FindService struct{}

func (s *FindService) Find(companyID string, dueDate time.Time) ([]domain.Invoice, error) {
	return []domain.Invoice{
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
	}, nil
}
