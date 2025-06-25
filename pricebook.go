package pricey

import (
	"time"
)

type ID = string

// Pricebook represents a pricebook entity.
type Pricebook struct {
	Id ID `json:"id" firestore:"id"`
	// OrgId identifier of the organization this pricebook belongs to.
	OrgId ID `json:"orgId" firestore:"orgId"`
	// GroupId identifier of the group this pricebook belongs to.
	GroupId ID `json:"groupId" firestore:"groupId"`
	// Name of the pricebook.
	Name string `json:"name" firestore:"name" validate:"required,min=2,max=100"`
	// Description of the pricebook.
	Description string `json:"description" firestore:"description" validate:"max=500"`
	// ImageId identifier associated with the pricebook.
	ImageId ID `json:"imageId" firestore:"imageId"`
	// Thumbnail identifier associated with the pricebook.
	ThumbnailId ID `json:"thumbnailId" firestore:"thumbnailId"`
	// Created timestamp when the pricebook was created.
	Created time.Time `json:"created" firestore:"created"`
	// Updated timestamp when the pricebook was last updated.
	Updated time.Time `json:"updated" firestore:"updated"`
	// Hidden flag indicating whether the pricebook is hidden or not.
	Hidden bool `json:"hidden" firestore:"hidden"`
}

type Category struct {
	Id ID `json:"id" firestore:"id"`
	// OrgId organization that owns this category
	OrgId ID `json:"orgId" firestore:"orgId"`
	// GroupId group that contains this category
	GroupId ID `json:"groupId" firestore:"groupId"`
	// ParentId parent category this category is nested under, if any
	ParentId *ID `json:"parentId" firestore:"parentId"`
	// PricebookId pricebook that contains this category
	PricebookId ID `json:"pricebookId" firestore:"pricebookId"`
	// Name human-readable name of the category
	Name string `json:"name" firestore:"name"`
	// Description optional brief description of the category
	Description string `json:"description" firestore:"description"`
	// HideFromCustomer flag to hide this category from customers when added to a quote or invoice
	HideFromCustomer bool `json:"hideFromCustomer" firestore:"hideFromCustomer"`
	// ImageId image associated with this category
	ImageId ID `json:"imageId" firestore:"imageId"`
	// ThumbnailId thumbnail associated with this category
	ThumbnailId ID `json:"thumbnailId" firestore:"thumbnailId"`
	// Created timestamp when the category was created
	Created time.Time `json:"created" firestore:"created"`
	// Updated timestamp when the category was last updated
	Updated time.Time `json:"updated" firestore:"updated"`
	// Hidden flag whether to hide this category
	Hidden bool `json:"hidden" firestore:"hidden"`
}

type Item struct {
	Id               ID        `json:"id" firestore:"id"`
	OrgId            ID        `json:"orgId" firestore:"orgId"`
	GroupId          ID        `json:"groupId" firestore:"groupId"`
	PricebookId      ID        `json:"pricebookId" firestore:"pricebookId"`
	CategoryId       ID        `json:"categoryId" firestore:"categoryId"`
	ParentIds        []ID      `json:"parentIds" firestore:"parentIds"`
	TagIds           []ID      `json:"tagIds" firestore:"tagIds"`
	SubItems         []SubItem `json:"subItems" firestore:"subItems"`
	Prices           []Price   `json:"prices" firestore:"prices"`
	Code             string    `json:"code" firestore:"code"`
	SKU              string    `json:"sku" firestore:"sku"`
	Name             string    `json:"name" firestore:"name"`
	Description      string    `json:"description" firestore:"description"`
	Cost             int       `json:"cost" firestore:"cost"`
	HideFromCustomer bool      `json:"hideFromCustomer" firestore:"hideFromCustomer"`
	ImageId          ID        `json:"imageId" firestore:"imageId"`
	ThumbnailId      ID        `json:"thumbnailId" firestore:"thumbnailId"`
	Created          time.Time `json:"created" firestore:"created"`
	Updated          time.Time `json:"updated" firestore:"updated"`
	Hidden           bool      `json:"hidden" firestore:"hidden"`
}

type SimpleItem struct {
	Id          ID     `json:"id" firestore:"id"`
	OrgId       ID     `json:"orgId" firestore:"orgId"`
	GroupId     ID     `json:"groupId" firestore:"groupId"`
	Name        string `json:"name" firestore:"name"`
	ThumbnailId ID     `json:"thumbnailId" firestore:"thumbnailId"`
}

type SubItem struct {
	SubItemID ID  `json:"subItemId" firestore:"subItemId"`
	PriceId   *ID `json:"priceId" firestore:"priceId"`
	Quantity  int `json:"quantity" firestore:"quantity"`
}

type Image struct {
	Id      ID        `json:"id" firestore:"id"`
	OrgId   ID        `json:"orgId" firestore:"orgId"`
	GroupId ID        `json:"groupId" firestore:"groupId"`
	Data    []byte    `json:"data" firestore:"data"`
	Base64  string    `json:"base64" firestore:"base64"`
	Url     string    `json:"url" firestore:"url"`
	Created time.Time `json:"created" firestore:"created"`
	Hidden  bool      `json:"hidden" firestore:"hidden"`
}

type Tag struct {
	Id              ID        `json:"id" firestore:"id"`
	OrgId           ID        `json:"orgId" firestore:"orgId"`
	GroupId         ID        `json:"groupId" firestore:"groupId"`
	PricebookId     ID        `json:"pricebookId" firestore:"pricebookId"`
	Name            string    `json:"name" firestore:"name"`
	Description     string    `json:"description" firestore:"description"`
	BackgroundColor string    `json:"backgroundColor" firestore:"backgroundColor"`
	TextColor       string    `json:"textColor" firestore:"textColor"`
	Created         time.Time `json:"created" firestore:"created"`
	Updated         time.Time `json:"updated" firestore:"updated"`
	Hidden          bool      `json:"hidden" firestore:"hidden"`
}

type Price struct {
	Id          ID        `json:"id" firestore:"id"`
	Name        string    `json:"name" firestore:"name"`
	Description string    `json:"description" firestore:"description"`
	Amount      int       `json:"amount" firestore:"amount"`
	Prefix      string    `json:"prefix" firestore:"prefix"`
	Suffix      string    `json:"suffix" firestore:"suffix"`
	Created     time.Time `json:"created" firestore:"created"`
	Updated     time.Time `json:"updated" firestore:"updated"`
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
