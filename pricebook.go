package pricey

import (
	"time"
)

type Pricebook struct {
	Id          int64
	Name        string
	Description string
	ImageId     int64
	ThumbnailId int64
	Created     time.Time
	Updated     time.Time
	Hidden      bool
}

type Category struct {
	Id               int64
	ParentId         *int64
	PricebookId      int64
	Name             string
	Description      string
	HideFromCustomer bool
	ImageId          int64
	ThumbnailId      int64
	Created          time.Time
	Updated          time.Time
	Hidden           bool
}

type Item struct {
	Id               int64
	PricebookId      int64
	CategoryId       int64
	ParentIds        []int64
	SubItemIds       []int64
	Code             string
	SKU              string
	Name             string
	Description      string
	Cost             float64
	PriceIds         []int64
	TagIds           []int64
	HideFromCustomer bool
	ImageId          int64
	ThumbnailId      int64
	Created          time.Time
	Updated          time.Time
	Hidden           bool
}

type SimpleItem struct {
	Id          int64
	Name        string
	ThumbnailId int64
}

type SubItem struct {
	SubItemId   int64
	ItemId      int64
	PricebookId int64
	Quantity    float64
}

type Image struct {
	Id          int64
	PricebookId int64
	Base64      string
	Url         string
	Created     time.Time
	Hidden      bool
}

type Tag struct {
	Id          int64
	PricebookId int64
	Name        string
	Description string
	Created     time.Time
	Updated     time.Time
	Hidden      bool
}

type Price struct {
	Id          int64
	ItemId      int64
	CategoryId  int64
	PricebookId int64
	Name        string
	Description string
	Amount      float64
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
