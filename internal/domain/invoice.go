package domain

import (
	"time"
)

type Invoice struct {
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
	Status    Status
}

type Status string

const (
	Unprocessed = Status("unprocessed")
	Processing  = Status("processing")
	Paid        = Status("paid")
	Error       = Status("error")
)

const (
	feeRate = 0.04
	taxRate = 0.1
)

func NewInvoice(issueDate, dueDate time.Time, amount int, status string) *Invoice {
	fee := int(float32(amount) * feeRate)
	tax := int(float32(fee) * taxRate)
	total := amount + fee + tax
	return &Invoice{
		IssueDate: issueDate,
		Amount:    amount,
		Fee:       fee,
		FeeRate:   feeRate,
		Tax:       tax,
		TaxRate:   taxRate,
		Total:     total,
		DueDate:   dueDate,
		Status:    Status(status),
	}
}
