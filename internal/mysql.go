package internal

import (
	"time"
)

type MySQL struct{}

var _ Selector = (*MySQL)(nil)

type Rows struct {
	Rows []Row
}

type Row struct {
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

func (s *MySQL) Select(companyID string, dueDate time.Time) (*Rows, error) {
	return &Rows{Rows: []Row{
		{
			IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
			Amount:    10000,
			Fee:       400,
			FeeRate:   0.04,
			Tax:       40,
			TaxRate:   0.10,
			DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
			Status:    "Processing",
		},
		{
			IssueDate: time.Date(1970, 1, 2, 9, 0, 0, 0, time.UTC),
			Amount:    5000,
			Fee:       200,
			FeeRate:   0.04,
			Tax:       20,
			TaxRate:   0.10,
			DueDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			Status:    "Processing",
		},
	}}, nil
}
