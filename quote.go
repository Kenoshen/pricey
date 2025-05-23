package pricey

import (
	"time"
)

type Quote struct {
	Id                     int        `json:"id"`
	Code                   string     `json:"code"`
	OrderNumber            string     `json:"orderNumber"`
	LogoId                 int        `json:"logoId"`
	PrimaryBackgroundColor string     `json:"primaryBackgroundColor"`
	PrimaryTextColor       string     `json:"primaryTextColor"`
	IssueDate              *time.Time `json:"issueDate"`
	ExpirationDate         *time.Time `json:"expirationDate"`
	PaymentTerms           string     `json:"paymentTerms"`
	Notes                  string     `json:"notes"`
	SenderId               int        `json:"senderId"`
	BillToId               int        `json:"billToId"`
	ShipToId               int        `json:"shipToId"`
	LineItemIds            []int      `json:"lineItemIds"`
	SubTotal               int        `json:"subTotal"`
	AdjustmentIds          []int      `json:"adjustmentsIds"`
	Total                  int        `json:"total"`
	BalanceDue             int        `json:"balanceDue"`
	BalancePercentDue      int        `json:"balancePercentDue"`
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
	Id              int       `json:"id"`
	QuoteId         int       `json:"quoteId"`
	ParentId        *int      `json:"parentId"`
	SubItemIds      []int     `json:"subItemIds"`
	ImageId         *int      `json:"imageId"`
	Description     string    `json:"description"`
	Quantity        int       `json:"quantity"`
	QuantitySuffix  string    `json:"quantitySuffix"`
	QuantityPrefix  string    `json:"quantityPrefix"`
	UnitPrice       int       `json:"unitPrice"`
	UnitPriceSuffix string    `json:"unitSuffix"`
	UnitPricePrefix string    `json:"unitPrefix"`
	Amount          *int      `json:"amount"`
	AmountSuffix    string    `json:"amountSuffix"`
	AmountPrefix    string    `json:"amountPrefix"`
	Open            bool      `json:"open"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
}

type Adjustment struct {
	Id          int            `json:"id"`
	QuoteId     int            `json:"quoteId"`
	Description string         `json:"description"`
	Type        AdjustmentType `json:"type"`
	Amount      int            `json:"amount"`
	Created     time.Time      `json:"created"`
	Updated     time.Time      `json:"updated"`
}

type AdjustmentType = int

const (
	AdjustmentTypeFlat    = 0
	AdjustmentTypePercent = 1
)

type Contact struct {
	Id          int
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
	Id                     int
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
	SubTotal               int
	Adjustments            []*Adjustment
	Total                  int
	BalanceDue             int
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
	Id               int
	Depth            int
	Number           string
	SubItems         []*PrintableLineItem
	Image            *Image
	Description      string
	Quantity         int
	QuantitySuffix   string
	QuantityPrefix   string
	UnitPrice        int
	UnitPriceSuffix  string
	UnitPricePrefix  string
	AmountOverridden bool
	Amount           int
	AmountSuffix     string
	AmountPrefix     string
	Created          time.Time
	Updated          time.Time
}
