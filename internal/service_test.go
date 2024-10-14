package internal

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestFindService_Find(t *testing.T) {
	tests := []struct {
		name    string
		rows    *Rows
		err     error
		want    []domain.Invoice
		wantErr error
	}{
		{
			name: "selector returns rows",
			rows: &Rows{Rows: []Row{
				{
					IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
					Amount:    10000,
					Fee:       400,
					FeeRate:   0.04,
					Tax:       40,
					TaxRate:   0.10,
					DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
					Status:    "unprocessed",
				},
				{
					IssueDate: time.Date(1970, 1, 2, 9, 0, 0, 0, time.UTC),
					Amount:    5000,
					Fee:       200,
					FeeRate:   0.04,
					Tax:       20,
					TaxRate:   0.10,
					DueDate:   time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
					Status:    "processing",
				},
			}},
			want: []domain.Invoice{
				{
					IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
					Amount:    10000,
					Fee:       400,
					FeeRate:   0.04,
					Tax:       40,
					TaxRate:   0.10,
					DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
					Status:    domain.Unprocessed,
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
		},
		{
			name:    "selector returns error",
			err:     errors.New("this is test"),
			wantErr: fmt.Errorf("find service error: %w", errors.New("this is test")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := SelectorFunc(func(string, time.Time) (*Rows, error) {
				return tt.rows, tt.err
			})
			s := FindService{Selector: selector}
			got, err := s.Find("1", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}