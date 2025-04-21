package pricey

import (
	"time"
)

type Quote struct {
	Id                     int64      `json:"id"`
	Code                   string     `json:"code"`
	OrderNumber            string     `json:"orderNumber"`
	LogoId                 int64      `json:"logoId"`
	PrimaryBackgroundColor string     `json:"primaryBackgroundColor"`
	PrimaryTextColor       string     `json:"primaryTextColor"`
	IssueDate              *time.Time `json:"issueDate"`
	ExpirationDate         *time.Time `json:"expirationDate"`
	PaymentTerms           string     `json:"paymentTerms"`
	Notes                  string     `json:"notes"`
	SenderId               int64      `json:"senderId"`
	BillToId               int64      `json:"billToId"`
	ShipToId               int64      `json:"shipToId"`
	LineItemIds            []int64    `json:"lineItemIds"`
	SubTotal               int64      `json:"subTotal"`
	AdjustmentIds          []int64    `json:"adjustmentsIds"`
	Total                  int64      `json:"total"`
	BalanceDue             int64      `json:"balanceDue"`
	BalancePercentDue      int64      `json:"balancePercentDue"`
	BalanceDueOn           *time.Time `json:"balanceDueOn"`
	PayUrl                 string     `json:"payUrl"`
	Sent                   bool       `json:"sent"`
	SentOn                 *time.Time `json:"sentOn"`
	Sold                   bool       `json:"sold"`
	SoldOn                 *time.Time `json:"soldOn"`
	Created                time.Time  `json:"created"`
	Updated                time.Time  `json:"updated"`
	Hidden                 bool       `json:"hidden"`
	Locked                 bool       `json:"locked"`
}

type LineItem struct {
	Id              int64     `json:"id"`
	QuoteId         int64     `json:"quoteId"`
	ParentId        *int64    `json:"parentId"`
	SubItemIds      []int64   `json:"subItemIds"`
	ImageId         *int64    `json:"imageId"`
	Description     string    `json:"description"`
	Quantity        int64     `json:"quantity"`
	QuantitySuffix  string    `json:"quantitySuffix"`
	QuantityPrefix  string    `json:"quantityPrefix"`
	UnitPrice       int64     `json:"unitPrice"`
	UnitPriceSuffix string    `json:"unitSuffix"`
	UnitPricePrefix string    `json:"unitPrefix"`
	Amount          *int64    `json:"amount"`
	AmountSuffix    string    `json:"amountSuffix"`
	AmountPrefix    string    `json:"amountPrefix"`
	Open            bool      `json:"open"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
}

type Adjustment struct {
	Id          int64          `json:"id"`
	QuoteId     int64          `json:"quoteId"`
	Description string         `json:"description"`
	Type        AdjustmentType `json:"type"`
	Amount      int64          `json:"amount"`
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

type PrintableQuote struct {
	Id                     int64
	Code                   string
	OrderNumber            string
	Logo                   *Image
	PrimaryBackgroundColor string
	PrimaryTextColor       string
	IssueDate              *time.Time
	ExpirationDate         *time.Time
	PaymentTerms           string
	Notes                  string
	Sender                 *Contact
	BillTo                 *Contact
	ShipTo                 *Contact
	LineItems              []*PrintableLineItem
	SubTotal               int64
	Adjustments            []*Adjustment
	Total                  int64
	BalanceDue             int64
	BalanceDueOn           *time.Time
	PayUrl                 string
	Sent                   bool
	SentOn                 *time.Time
	Sold                   bool
	SoldOn                 *time.Time
	Created                time.Time
	Updated                time.Time
	Hidden                 bool
	Locked                 bool
}

type PrintableLineItem struct {
	Id               int64
	Depth            int
	Number           string
	SubItems         []*PrintableLineItem
	Image            *Image
	Description      string
	Quantity         int64
	QuantitySuffix   string
	QuantityPrefix   string
	UnitPrice        int64
	UnitPriceSuffix  string
	UnitPricePrefix  string
	AmountOverridden bool
	Amount           int64
	AmountSuffix     string
	AmountPrefix     string
	Created          time.Time
	Updated          time.Time
}
