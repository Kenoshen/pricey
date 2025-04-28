package pricey

import (
	"time"
)

type Pricebook struct {
	Id          int
	Name        string
	Description string
	ImageId     int
	ThumbnailId int
	Created     time.Time
	Updated     time.Time
	Hidden      bool
}

type Category struct {
	Id               int
	ParentId         *int
	PricebookId      int
	Name             string
	Description      string
	HideFromCustomer bool
	ImageId          int
	ThumbnailId      int
	Created          time.Time
	Updated          time.Time
	Hidden           bool
}

type Item struct {
	Id               int
	PricebookId      int
	CategoryId       int
	ParentIds        []int
	SubItemIds       []SubItem
	Code             string
	SKU              string
	Name             string
	Description      string
	Cost             int
	PriceIds         []int
	TagIds           []int
	HideFromCustomer bool
	ImageId          int
	ThumbnailId      int
	Created          time.Time
	Updated          time.Time
	Hidden           bool
}

type SimpleItem struct {
	Id          int
	Name        string
	ThumbnailId int
}

type SubItem struct {
	SubItemId int
	Quantity  int
}

type Image struct {
	Id          int
	PricebookId int
	Data        []byte
	Base64      string
	Url         string
	Created     time.Time
	Hidden      bool
}

type Tag struct {
	Id              int
	PricebookId     int
	Name            string
	Description     string
	BackgroundColor string
	TextColor       string
	Created         time.Time
	Updated         time.Time
	Hidden          bool
}

type Price struct {
	Id          int
	ItemId      int
	CategoryId  int
	PricebookId int
	Name        string
	Description string
	Amount      int
	Prefix      string
	Suffix      string
	Created     time.Time
	Updated     time.Time
	Hidden      bool
}

/*
Keywords and Ideas for the future:

Add ons
Service
Task
Widget
Part
Equipment
Pricing
Promotion
Adjustment
Variation
Unit
Cost
Customer
Membership
Logo
Technician
Dispatcher
Salesman
CustomerPortal
LaborRate
UnitPrice
GoodBetterBest
Markup
OverheadRate
Tax
Job
JobType
SKU
Collection
Brand
Manufacturer
Template
Signature
PaymentPage
Grid
List
Leftover
Inventory
Keyword
Labor
Extra
Materials
Fee
ItemGroup
Group
Code
BulkUpdate
Catalogs
Throwaway
Order
PartsList
Supplier
Provider
Recommendation
Member
MaterialCost
Model
Upgrade
ProductFamily
Currency
Warranty
Proposal
*/
