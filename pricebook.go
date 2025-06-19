package pricey

import (
	"time"
)

type ID = string

// Pricebook represents a pricebook entity.
type Pricebook struct {
	Id ID `json:"id"`
	// OrgId identifier of the organization this pricebook belongs to.
	OrgId ID `json:"orgId"`
	// GroupId identifier of the group this pricebook belongs to.
	GroupId ID `json:"groupId"`
	// Name of the pricebook.
	Name string `json:"name" validate:"required,min=2,max=100"`
	// Description of the pricebook.
	Description string `json:"description" validate:"max=500"`
	// ImageId identifier associated with the pricebook.
	ImageId ID `json:"imageId"`
	// Thumbnail identifier associated with the pricebook.
	ThumbnailId ID `json:"thumbnailId"`
	// Created timestamp when the pricebook was created.
	Created time.Time `json:"created"`
	// Updated timestamp when the pricebook was last updated.
	Updated time.Time `json:"updated"`
	// Hidden flag indicating whether the pricebook is hidden or not.
	Hidden bool `json:"hidden"`
}

type Category struct {
	Id ID `json:"id"`
	// OrgId organization that owns this category
	OrgId ID `json:"orgId"`
	// GroupId group that contains this category
	GroupId ID `json:"groupId"`
	// ParentId parent category this category is nested under, if any
	ParentId *ID `json:"parentId"`
	// PricebookId pricebook that contains this category
	PricebookId ID `json:"pricebookId"`
	// Name human-readable name of the category
	Name string `json:"name"`
	// Description optional brief description of the category
	Description string `json:"description"`
	// HideFromCustomer flag to hide this category from customers when added to a quote or invoice
	HideFromCustomer bool `json:"hideFromCustomer"`
	// ImageId image associated with this category
	ImageId ID `json:"imageId"`
	// ThumbnailId thumbnail associated with this category
	ThumbnailId ID `json:"thumbnailId"`
	// Created timestamp when the category was created
	Created time.Time `json:"created"`
	// Updated timestamp when the category was last updated
	Updated time.Time `json:"updated"`
	// Hidden flag whether to hide this category
	Hidden bool `json:"hidden"`
}

type Item struct {
	Id               ID        `json:"id"`
	OrgId            ID        `json:"orgId"`
	GroupId          ID        `json:"groupId"`
	PricebookId      ID        `json:"pricebookId"`
	CategoryId       ID        `json:"categoryId"`
	ParentIds        []ID      `json:"parentIds"`
	SubItemIds       []SubItem `json:"subItemIds"`
	Code             string    `json:"code"`
	SKU              string    `json:"sku"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Cost             int       `json:"cost"`
	PriceIds         []ID      `json:"priceIds"`
	TagIds           []ID      `json:"tagIds"`
	HideFromCustomer bool      `json:"hideFromCustomer"`
	ImageId          ID        `json:"imageId"`
	ThumbnailId      ID        `json:"thumbnailId"`
	Created          time.Time `json:"created"`
	Updated          time.Time `json:"updated"`
	Hidden           bool      `json:"hidden"`
}

type SimpleItem struct {
	Id          ID     `json:"id"`
	OrgId       ID     `json:"orgId"`
	GroupId     ID     `json:"groupId"`
	Name        string `json:"name"`
	ThumbnailId ID     `json:"thumbnailId"`
}

type SubItem struct {
	SubItemID ID  `json:"subItemId"`
	PriceId   *ID `json:"priceId"`
	Quantity  int `json:"quantity"`
}

type Image struct {
	Id      ID        `json:"id"`
	OrgId   ID        `json:"orgId"`
	GroupId ID        `json:"groupId"`
	Data    []byte    `json:"data"`
	Base64  string    `json:"base64"`
	Url     string    `json:"url"`
	Created time.Time `json:"created"`
	Hidden  bool      `json:"hidden"`
}

type Tag struct {
	Id              ID        `json:"id"`
	OrgId           ID        `json:"orgId"`
	GroupId         ID        `json:"groupId"`
	PricebookId     ID        `json:"pricebookId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	BackgroundColor string    `json:"backgroundColor"`
	TextColor       string    `json:"textColor"`
	Created         time.Time `json:"created"`
	Updated         time.Time `json:"updated"`
	Hidden          bool      `json:"hidden"`
}

type Price struct {
	Id          ID        `json:"id"`
	OrgId       ID        `json:"orgId"`
	GroupId     ID        `json:"groupId"`
	ItemId      ID        `json:"itemId"`
	CategoryId  ID        `json:"categoryId"`
	PricebookId ID        `json:"pricebookId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Amount      int       `json:"amount"`
	Prefix      string    `json:"prefix"`
	Suffix      string    `json:"suffix"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	Hidden      bool      `json:"hidden"`
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
