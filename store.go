package pricey

import (
	"context"
)

type Store interface {
  // ////////////
  // PRICEBOOK
  // ////////////

	CreatePricebook(ctx context.Context, name, description string) (*Pricebook, error)
	GetPricebook(ctx context.Context, id int64) (*Pricebook, error)
	GetPricebooks(ctx context.Context) ([]*Pricebook, error)
	UpdatePricebook(ctx context.Context, pb Pricebook) (*Pricebook, error)
	DeletePricebook(ctx context.Context, id int64) error

  // ////////////
  // CATEGORY
  // ////////////

	CreateCategory(ctx context.Context, name, description string) (*Category, error)
	GetCategory(ctx context.Context, id int64) (*Category, error)
	GetCategories(ctx context.Context, pricebookId int64) ([]*Category, error)
	UpdateCategoryInfo(ctx context.Context, id int64, name, description string) (*Category, error)
	UpdateCategoryImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Category, error)
	MoveCategory(ctx context.Context, id int64, pricebookId int64, parentId *int64) (*Category, error)
	DeleteCategory(ctx context.Context, id int64) error
	SearchCategories(ctx context.Context, search string) error

  // ////////////
  // ITEMS
  // ////////////

	CreateItem(ctx context.Context, categoryId int64, name, description string) (*Item, error)
	GetItem(ctx context.Context, id int64) (*Item, error)
	GetSimpleItem(ctx context.Context, id int64) (*SimpleItem, error)
	GetItemsInCategory(ctx context.Context, categoryId int64) ([]*Item, error)
	MoveItem(ctx context.Context, id int64, newCategoryId int64) (*Item, error)
	AddSubItem(ctx context.Context, id int64, subItemId int64, quantity int64) (*Item, error)
	UpdateSubItemQuantity(ctx context.Context, id int64, subItemId int64, quantity int64) (*Item, error)
	RemoveSubItem(ctx context.Context, id int64, subItemId int64) (*Item, error)
	UpdateItemInfo(ctx context.Context, id int64, code, sku, name, description string) (*Item, error)
	UpdateItemCost(ctx context.Context, id int64, cost float64) (*Item, error)
	AddItemPrice(ctx context.Context, id int64, priceId int64) (*Item, error)
	RemoveItemPrice(ctx context.Context, id int64, priceId int64) (*Item, error)
	AddItemTag(ctx context.Context, id int64, tagId int64) (*Item, error)
	RemoveItemTag(ctx context.Context, id int64, tagId int64) (*Item, error)
	UpdateItemHideFromCustomer(ctx context.Context, id int64, hideFromCustomer bool) (*Item, error)
	UpdateItemImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Item, error)
	DeleteItem(ctx context.Context, id int64) error
	SearchItems(ctx context.Context, search string) ([]*Item, error)

  // ////////////
  // PRICE
  // ////////////

	CreatePrice(ctx context.Context, amount float64) (*Price, error)
	GetPrice(ctx context.Context, id int64) (*Price, error)
	UpdatePrice(ctx context.Context, p Price) (*Price, error)
	DeletePrice(ctx context.Context, id int64) error

  // ////////////
  // TAG
  // ////////////

	CreateTag(ctx context.Context, name, description string) (*Tag, error)
	GetTag(ctx context.Context, id int64) (*Tag, error)
	UpdateTagName(ctx context.Context, id int64, name string) (*Tag, error)
	UpdateTagDescription(ctx context.Context, id int64, description string) (*Tag, error)
	DeleteTag(ctx context.Context, id int64) error

  // ////////////
  // IMAGE
  // ////////////

	CreateImage(ctx context.Context, data []byte) (int64, error)
	GetImageUrl(ctx context.Context, id int64) (string, error)
	GetImageBase64(ctx context.Context, id int64) (string, error)
	GetImageData(ctx context.Context, id int64) ([]byte, error)
	DeleteImage(ctx context.Context, id int64) error

  // ////////////
  // QUOTE
  // ////////////

	CreateQuote(ctx context.Context) (*Quote, error)
	CreateQuoteFromTemplate(ctx context.Context, templateId int64) (*Quote, error)
	CreateDuplicateQuote(ctx context.Context, quoteId int64) (*Quote, error)
	GetQuote(ctx context.Context, id int64) (*Quote, error)

  // ////////////
  // LINE ITEM
  // ////////////

	CreateLineItem(ctx context.Context, quoteId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error)
	CreateSubLineItem(ctx context.Context, quoteId, parentId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error)
	CreateDuplicateLineItem(ctx context.Context, id int64) (*LineItem, error)
	GetLineItem(ctx context.Context, id int64) (*LineItem, error)
	MoveLineItem(ctx context.Context, id int64, parentId *int64) (*LineItem, error)
	UpdateLineItemImage(ctx context.Context, id int64, imageId *int64) (*LineItem, error)
	UpdateLineItemDescription(ctx context.Context, id int64, description string) (*LineItem, error)
	UpdateLineItemQuantity(ctx context.Context, id int64, quantity float64, prefix, suffix string) (*LineItem, error)
	UpdateLineItemUnitPrice(ctx context.Context, id int64, unitPrice float64, prefix, suffix string) (*LineItem, error)
	UpdateLineItemAmount(ctx context.Context, id int64, amount *float64, prefix, suffix string) (*LineItem, error)
	UpdateLineItemOpen(ctx context.Context, id int64, open bool) (*LineItem, error)
	DeleteLineItem(ctx context.Context, id int64) error

  // ////////////
  // ADJUSTMENT
  // ////////////

	AddAdjustment(ctx context.Context, quoteId int64, description string, amount float64, adjustmentType AdjustmentType) (*LineItem, error)
	UpdateAdjustment(ctx context.Context, id int64, description string, amount float64, adjustmentType AdjustmentType) (*LineItem, error)
	RemoveAdjustment(ctx context.Context, id int64) (*LineItem, error)

  // ////////////
  // CONTACT
  // ////////////

	GetContact(ctx context.Context, id int64) (*Contact, error)
}
