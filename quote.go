package pricey

import (
	"time"
)

type Quote struct {
	Id                     ID         `json:"id" firestore:"id" firestore:"id"`
	Code                   string     `json:"code" firestore:"code" firestore:"code"`
	OrderNumber            string     `json:"orderNumber" firestore:"orderNumber" firestore:"orderNumber"`
	LogoId                 ID         `json:"logoId" firestore:"logoId" firestore:"logoId"`
	PrimaryBackgroundColor string     `json:"primaryBackgroundColor" firestore:"primaryBackgroundColor" firestore:"primaryBackgroundColor"`
	PrimaryTextColor       string     `json:"primaryTextColor" firestore:"primaryTextColor" firestore:"primaryTextColor"`
	IssueDate              *time.Time `json:"issueDate" firestore:"issueDate" firestore:"issueDate"`
	ExpirationDate         *time.Time `json:"expirationDate" firestore:"expirationDate" firestore:"expirationDate"`
	PaymentTerms           string     `json:"paymentTerms" firestore:"paymentTerms" firestore:"paymentTerms"`
	Notes                  string     `json:"notes" firestore:"notes" firestore:"notes"`
	SenderId               ID         `json:"senderId" firestore:"senderId" firestore:"senderId"`
	BillToId               ID         `json:"billToId" firestore:"billToId" firestore:"billToId"`
	ShipToId               ID         `json:"shipToId" firestore:"shipToId" firestore:"shipToId"`
	LineItemIds            []ID       `json:"lineItemIds" firestore:"lineItemIds" firestore:"lineItemIds"`
	SubTotal               int        `json:"subTotal" firestore:"subTotal" firestore:"subTotal"`
	AdjustmentIds          []ID       `json:"adjustmentsIds" firestore:"adjustmentsIds" firestore:"adjustmentsIds"`
	Total                  int        `json:"total" firestore:"total" firestore:"total"`
	BalanceDue             int        `json:"balanceDue" firestore:"balanceDue" firestore:"balanceDue"`
	BalancePercentDue      int        `json:"balancePercentDue" firestore:"balancePercentDue" firestore:"balancePercentDue"`
	BalanceDueOn           *time.Time `json:"balanceDueOn" firestore:"balanceDueOn" firestore:"balanceDueOn"`
	PayUrl                 string     `json:"payUrl" firestore:"payUrl" firestore:"payUrl"`
	Sent                   bool       `json:"sent" firestore:"sent" firestore:"sent"`
	SentOn                 *time.Time `json:"sentOn" firestore:"sentOn" firestore:"sentOn"`
	Sold                   bool       `json:"sold" firestore:"sold" firestore:"sold"`
	SoldOn                 *time.Time `json:"soldOn" firestore:"soldOn" firestore:"soldOn"`
	Created                time.Time  `json:"created" firestore:"created" firestore:"created"`
	Updated                time.Time  `json:"updated" firestore:"updated" firestore:"updated"`
	Hidden                 bool       `json:"hidden" firestore:"hidden" firestore:"hidden"`
	Locked                 bool       `json:"locked" firestore:"locked" firestore:"locked"`
}

type LineItem struct {
	Id              ID        `json:"id" firestore:"id" firestore:"id"`
	QuoteId         ID        `json:"quoteId" firestore:"quoteId" firestore:"quoteId"`
	ParentId        *ID       `json:"parentId" firestore:"parentId" firestore:"parentId"`
	SubItemIds      []ID      `json:"subItemIds" firestore:"subItemIds" firestore:"subItemIds"`
	ImageId         *ID       `json:"imageId" firestore:"imageId" firestore:"imageId"`
	Description     string    `json:"description" firestore:"description" firestore:"description"`
	Quantity        int       `json:"quantity" firestore:"quantity" firestore:"quantity"`
	QuantitySuffix  string    `json:"quantitySuffix" firestore:"quantitySuffix" firestore:"quantitySuffix"`
	QuantityPrefix  string    `json:"quantityPrefix" firestore:"quantityPrefix" firestore:"quantityPrefix"`
	UnitPrice       int       `json:"unitPrice" firestore:"unitPrice" firestore:"unitPrice"`
	UnitPriceSuffix string    `json:"unitSuffix" firestore:"unitSuffix" firestore:"unitSuffix"`
	UnitPricePrefix string    `json:"unitPrefix" firestore:"unitPrefix" firestore:"unitPrefix"`
	Amount          *int      `json:"amount" firestore:"amount" firestore:"amount"`
	AmountSuffix    string    `json:"amountSuffix" firestore:"amountSuffix" firestore:"amountSuffix"`
	AmountPrefix    string    `json:"amountPrefix" firestore:"amountPrefix" firestore:"amountPrefix"`
	Open            bool      `json:"open" firestore:"open" firestore:"open"`
	Created         time.Time `json:"created" firestore:"created" firestore:"created"`
	Updated         time.Time `json:"updated" firestore:"updated" firestore:"updated"`
}

type Adjustment struct {
	Id          ID             `json:"id" firestore:"id" firestore:"id"`
	QuoteId     ID             `json:"quoteId" firestore:"quoteId" firestore:"quoteId"`
	Description string         `json:"description" firestore:"description" firestore:"description"`
	Type        AdjustmentType `json:"type" firestore:"type" firestore:"type"`
	Amount      int            `json:"amount" firestore:"amount" firestore:"amount"`
	Created     time.Time      `json:"created" firestore:"created" firestore:"created"`
	Updated     time.Time      `json:"updated" firestore:"updated" firestore:"updated"`
}

type AdjustmentType = int

const (
	AdjustmentTypeFlat    = 0
	AdjustmentTypePercent = 1
)

// Contact represents a contact in the company.
type Contact struct {
	Id          ID       `json:"id" firestore:"id"`
	Name        string   `json:"name" firestore:"name"`
	CompanyName string   `json:"companyName" firestore:"companyName"`
	Phones      []string `json:"phones" firestore:"phones"`
	Emails      []string `json:"emails" firestore:"emails"`
	Websites    []string `json:"websites" firestore:"websites"`
	Street      string   `json:"street" firestore:"street"`
	City        string   `json:"city" firestore:"city"`
	State       string   `json:"state" firestore:"state"`
	Zip         string   `json:"zip" firestore:"zip"`
}

// PrintableQuote represents a printable quote.
type PrintableQuote struct {
	Id                     ID                   `json:"id" firestore:"id"`
	Code                   string               `json:"code" firestore:"code"`
	OrderNumber            string               `json:"orderNumber" firestore:"orderNumber"`
	Logo                   *Image               `json:"logo" firestore:"logo"`
	PrimaryBackgroundColor string               `json:"primaryBackgroundColor" firestore:"primaryBackgroundColor"`
	PrimaryTextColor       string               `json:"primaryTextColor" firestore:"primaryTextColor"`
	IssueDate              *time.Time           `json:"issueDate" firestore:"issueDate"`
	ExpirationDate         *time.Time           `json:"expirationDate" firestore:"expirationDate"`
	PaymentTerms           string               `json:"paymentTerms" firestore:"paymentTerms"`
	Notes                  string               `json:"notes" firestore:"notes"`
	Sender                 *Contact             `json:"sender" firestore:"sender"`
	BillTo                 *Contact             `json:"billTo" firestore:"billTo"`
	ShipTo                 *Contact             `json:"shipTo" firestore:"shipTo"`
	LineItems              []*PrintableLineItem `json:"lineItems" firestore:"lineItems"`
	SubTotal               int                  `json:"subTotal" firestore:"subTotal"`
	Adjustments            []*Adjustment        `json:"adjustments" firestore:"adjustments"`
	Total                  int                  `json:"total" firestore:"total"`
	BalanceDue             int                  `json:"balanceDue" firestore:"balanceDue"`
	BalanceDueOn           *time.Time           `json:"balanceDueOn" firestore:"balanceDueOn"`
	PayUrl                 string               `json:"payUrl" firestore:"payUrl"`
	Sent                   bool                 `json:"sent" firestore:"sent"`
	SentOn                 *time.Time           `json:"sentOn" firestore:"sentOn"`
	Sold                   bool                 `json:"sold" firestore:"sold"`
	SoldOn                 *time.Time           `json:"soldOn" firestore:"soldOn"`
	Created                time.Time            `json:"created" firestore:"created"`
	Updated                time.Time            `json:"updated" firestore:"updated"`
	Hidden                 bool                 `json:"hidden" firestore:"hidden"`
	Locked                 bool                 `json:"locked" firestore:"locked"`
}

// PrintableLineItem represents a line item in the printable quote.
type PrintableLineItem struct {
	Id               ID                   `json:"id" firestore:"id"`
	Depth            int                  `json:"depth" firestore:"depth"`
	Number           string               `json:"number" firestore:"number"`
	SubItems         []*PrintableLineItem `json:"subItems" firestore:"subItems"`
	Image            *Image               `json:"image" firestore:"image"`
	Description      string               `json:"description" firestore:"description"`
	Quantity         int                  `json:"quantity" firestore:"quantity"`
	QuantitySuffix   string               `json:"quantitySuffix" firestore:"quantitySuffix"`
	QuantityPrefix   string               `json:"quantityPrefix" firestore:"quantityPrefix"`
	UnitPrice        int                  `json:"unitPrice" firestore:"unitPrice"`
	UnitPriceSuffix  string               `json:"unitPriceSuffix" firestore:"unitPriceSuffix"`
	UnitPricePrefix  string               `json:"unitPricePrefix" firestore:"unitPricePrefix"`
	AmountOverridden bool                 `json:"amountOverridden" firestore:"amountOverridden"`
	Amount           int                  `json:"amount" firestore:"amount"`
	AmountSuffix     string               `json:"amountSuffix" firestore:"amountSuffix"`
	AmountPrefix     string               `json:"amountPrefix" firestore:"amountPrefix"`
	Created          time.Time            `json:"created" firestore:"created"`
	Updated          time.Time            `json:"updated" firestore:"updated"`
}
