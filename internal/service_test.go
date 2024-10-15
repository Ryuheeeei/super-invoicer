package internal

import (
	"context"
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
					Total:     10440,
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
					Total:     5220,
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
					Total:     10440,
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
					Total:     5220,
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
			selector := SelectorFunc(func(context.Context, string, time.Time) (*Rows, error) {
				return tt.rows, tt.err
			})
			s := FindService{Selector: selector}
			got, err := s.Find(context.Background(), "1", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC))
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegisterService_Insert(t *testing.T) {
	type args struct {
		companyID string
		issueDate time.Time
		amount    int
		dueDate   time.Time
		status    string
	}
	tests := []struct {
		name     string
		args     args
		inserter Inserter
		want     *domain.Invoice
		wantErr  error
	}{
		{
			name: "inserter returns inserted row",
			args: args{companyID: "1", issueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC), amount: 10000, dueDate: time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC), status: "unprocessed"},
			inserter: InserterFunc(func(ctx context.Context, companyID string, invoice *domain.Invoice) (*Row, error) {
				assert.Equal(t, "1", companyID)
				assert.Equal(t, &domain.Invoice{
					IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
					Amount:    10000,
					Fee:       400,
					FeeRate:   0.04,
					Tax:       40,
					TaxRate:   0.10,
					Total:     10440,
					DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
					Status:    domain.Unprocessed,
				}, invoice)
				return &Row{
					IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
					Amount:    10000,
					Fee:       400,
					FeeRate:   0.04,
					Tax:       40,
					TaxRate:   0.10,
					Total:     10440,
					DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
					Status:    "unprocessed",
				}, nil
			}),
			want: &domain.Invoice{
				IssueDate: time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC),
				Amount:    10000,
				Fee:       400,
				FeeRate:   0.04,
				Tax:       40,
				TaxRate:   0.10,
				Total:     10440,
				DueDate:   time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC),
				Status:    domain.Unprocessed,
			},
		},
		{
			name: "inserter returns error",
			inserter: InserterFunc(func(ctx context.Context, s string, i *domain.Invoice) (*Row, error) {
				return nil, errors.New("this is test")
			}),
			wantErr: fmt.Errorf("insert error: %w", errors.New("this is test")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := RegisterService{Inserter: tt.inserter}
			got, err := s.Register(context.Background(), "1", time.Date(1970, 1, 1, 9, 0, 0, 0, time.UTC), 10000, time.Date(2024, 10, 30, 0, 0, 0, 0, time.UTC), "unprocessed")
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
