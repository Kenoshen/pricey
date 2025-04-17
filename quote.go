package pricey

import (
	"time"
)

type Quote struct {
	Id                int64      `json:"id"`
	Code              string     `json:"code"`
	OrderNumber       string     `json:"orderNumber"`
	LogoId            int64      `json:"logoId"`
	IssueDate         *time.Time `json:"issueDate"`
	ExpirationDate    *time.Time `json:"expirationDate"`
	PaymentTerms      string     `json:"paymentTerms"`
	Notes             string     `json:"notes"`
	SenderId          int64      `json:"senderId"`
	BillToId          int64      `json:"billToId"`
	ShipToId          int64      `json:"shipToId"`
	LineItemIds       []int64    `json:"lineItemIds"`
	SubTotal          float64    `json:"subTotal"`
	AdjustmentIds     []int64    `json:"adjustmentsIds"`
	Total             float64    `json:"total"`
	BalanceDue        float64    `json:"balanceDue"`
	BalancePercentDue float64    `json:"balancePercentDue"`
	BalanceDueOn      *time.Time `json:"balanceDueOn"`
	PayUrl            string     `json:"payUrl"`
	Sent              bool       `json:"sent"`
	SentOn            *time.Time `json:"sentOn"`
	Sold              bool       `json:"sold"`
	SoldOn            *time.Time `json:"soldOn"`
	Created           time.Time  `json:"created"`
	Updated           time.Time  `json:"updated"`
	Hidden            bool       `json:"hidden"`
	Locked            bool       `json:"locked"`
}

type LineItem struct {
	Id             int64     `json:"id"`
	QuoteId        int64     `json:"quoteId"`
	ParentId       *int64    `json:"parentId"`
	ImageId        *int64    `json:"imageId"`
	Description    string    `json:"description"`
	Quantity       float64   `json:"quantity"`
	QuantitySuffix string    `json:"quantitySuffix"`
	QuantityPrefix string    `json:"quantityPrefix"`
	UnitPrice      float64   `json:"unitPrice"`
	UnitSuffix     string    `json:"unitSuffix"`
	UnitPrefix     string    `json:"unitPrefix"`
	Amount         *float64  `json:"amount"`
	AmountSuffix   string    `json:"amountSuffix"`
	AmountPrefix   string    `json:"amountPrefix"`
	Open           bool      `json:"open"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}

type Adjustment struct {
	Id          int64          `json:"id"`
	QuoteId     int64          `json:"quoteId"`
	Description string         `json:"description"`
	Type        AdjustmentType `json:"type"`
	Amount      float64        `json:"amount"`
	Created     time.Time      `json:"created"`
	Updated     time.Time      `json:"updated"`
}

type AdjustmentType = int64

const (
	AdjustmentTypeFlat    = 0
	AdjustmentTypePercent = 1
)

type Contact struct {
	Id          int64
	Name        string
	CompanyName string
	Phones      []string
	Emails      []string
	Websites    []string
	Street      string
	City        string
	State       string
	Zip         string
}

type FullQuote struct {
	Id                int64
	Code              string
	OrderNumber       string
	Logo              *Image
	IssueDate         *time.Time
	ExpirationDate    *time.Time
	PaymentTerms      string
	Notes             string
	Sender            *Contact
	BillTo            *Contact
	ShipTo            *Contact
	LineItems         []*FullLineItem
	SubTotal          float64
	Adjustments       []*Adjustment
	Total             float64
	BalanceDue        float64
	BalancePercentDue float64
	BalanceDueOn      *time.Time
	PayUrl            string
	Sent              bool
	SentOn            *time.Time
	Sold              bool
	SoldOn            *time.Time
	Created           time.Time
	Updated           time.Time
	Hidden            bool
	Locked            bool
}

type FullLineItem struct {
	Id             int64
	SubItems       []*FullLineItem
	Image          *Image
	Description    string
	Quantity       float64
	QuantitySuffix string
	QuantityPrefix string
	UnitPrice      float64
	UnitSuffix     string
	UnitPrefix     string
	Amount         *float64
	AmountSuffix   string
	AmountPrefix   string
	Open           bool
	Created        time.Time
	Updated        time.Time
}
