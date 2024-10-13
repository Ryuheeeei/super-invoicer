package domain

import "time"

type Invoice struct {
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
