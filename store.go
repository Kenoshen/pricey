package pricey

import (
	"context"
	"time"
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
	RecoverPricebook(ctx context.Context, id int64) error

	// ////////////
	// CATEGORY
	// ////////////

	CreateCategory(ctx context.Context, name, description string) (*Category, error)
	GetCategory(ctx context.Context, id int64) (*Category, error)
	GetCategories(ctx context.Context, pricebookId int64) ([]*Category, error)
	UpdateCategoryInfo(ctx context.Context, id int64, name, description string) (*Category, error)
	UpdateCategoryImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Category, error)
	MoveCategory(ctx context.Context, id int64, parentId int64) (*Category, error)
	DeleteCategory(ctx context.Context, id int64) error
	DeletePricebookCategories(ctx context.Context, pricebookId int64) error
	RecoverCategory(ctx context.Context, id int64) error
	RecoverPricebookCategories(ctx context.Context, pricebookId int64) error

	// ////////////
	// ITEMS
	// ////////////

	CreateItem(ctx context.Context, categoryId int64, name, description string) (*Item, error)
	GetItem(ctx context.Context, id int64) (*Item, error)
	GetSimpleItem(ctx context.Context, id int64) (*SimpleItem, error)
	GetItemsInCategory(ctx context.Context, categoryId int64) ([]*Item, error)
	MoveItem(ctx context.Context, id int64, newCategoryId int64) (*Item, error)
	UpdateItemInfo(ctx context.Context, id int64, code, sku, name, description string) (*Item, error)
	UpdateItemCost(ctx context.Context, id int64, cost float64) (*Item, error)
	AddItemPrice(ctx context.Context, id int64, priceId int64) (*Item, error)
	RemoveItemPrice(ctx context.Context, id int64, priceId int64) (*Item, error)
	AddItemTag(ctx context.Context, id int64, tagId int64) (*Item, error)
	RemoveItemTag(ctx context.Context, id int64, tagId int64) (*Item, error)
	RemoveTagFromItems(ctx context.Context, pricebookId, tagId int64) error
	UpdateItemHideFromCustomer(ctx context.Context, id int64, hideFromCustomer bool) (*Item, error)
	UpdateItemImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Item, error)
	SearchItemsInPricebook(ctx context.Context, pricebookId int64, search string) ([]*Item, error)
	DeleteItem(ctx context.Context, id int64) error
	DeleteCategoryItems(ctx context.Context, categoryId int64) error
	DeletePricebookItems(ctx context.Context, pricebookId int64) error
	RecoverItem(ctx context.Context, id int64) error
	RecoverCategoryItems(ctx context.Context, categoryId int64) error
	RecoverPricebookItems(ctx context.Context, pricebookId int64) error

	// ////////////
	// SUB ITEM
	// ////////////

	AddSubItem(ctx context.Context, id int64, subItemId int64, quantity int64) (*Item, error)
	UpdateSubItemQuantity(ctx context.Context, id int64, subItemId int64, quantity int64) (*Item, error)
	RemoveSubItem(ctx context.Context, id int64, subItemId int64) (*Item, error)

	// ////////////
	// PRICE
	// ////////////

	CreatePrice(ctx context.Context, itemId int64, amount float64) (*Price, error)
	GetPrice(ctx context.Context, id int64) (*Price, error)
	GetPricesByItem(ctx context.Context, itemId int64) ([]*Price, error)
	MovePricesByItem(ctx context.Context, itemId, categoryId int64) error
	UpdatePrice(ctx context.Context, p Price) (*Price, error)
	DeletePrice(ctx context.Context, id int64) error
	DeletePricesByItem(ctx context.Context, itemId int64) error
	DeleteCategoryPrices(ctx context.Context, categoryId int64) error
	DeletePricebookPrices(ctx context.Context, pricebookId int64) error
	RecoverPricesByItem(ctx context.Context, itemId int64) error
	RecoverCategoryPrices(ctx context.Context, categoryId int64) error
	RecoverPricebookPrices(ctx context.Context, pricebookId int64) error

	// ////////////
	// TAG
	// ////////////

	CreateTag(ctx context.Context, pricebookId int64, name, description string) (*Tag, error)
	GetTag(ctx context.Context, id int64) (*Tag, error)
	GetTags(ctx context.Context, pricebookId int64) ([]*Tag, error)
	UpdateTagInfo(ctx context.Context, id int64, name, description string) (*Tag, error)
	SearchTags(ctx context.Context, pricebookId int64, search string) ([]*Tag, error)
	DeleteTag(ctx context.Context, id int64) error
	DeletePricebookTags(ctx context.Context, id int64) error
	RecoverPricebookTags(ctx context.Context, id int64) error

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
	CreateDuplicateQuote(ctx context.Context, quoteId int64) (*Quote, error)
	GetQuote(ctx context.Context, id int64) (*Quote, error)
	UpdateQuoteCode(ctx context.Context, id int64, code string) (*Quote, error)
	UpdateQuoteOrderNumber(ctx context.Context, id int64, orderNumber string) (*Quote, error)
	UpdateQuoteLogoId(ctx context.Context, id int64, logoId int64) (*Quote, error)
	UpdateQuoteIssueDate(ctx context.Context, id int64, issueDate *time.Time) (*Quote, error)
	UpdateQuoteExpirationDate(ctx context.Context, id int64, expirationDate *time.Time) (*Quote, error)
	UpdateQuotePaymentTerms(ctx context.Context, id int64, paymentTerms string) (*Quote, error)
	UpdateQuoteNotes(ctx context.Context, id int64, notes string) (*Quote, error)
	UpdateQuoteSenderId(ctx context.Context, id int64, contactId int64) (*Quote, error)
	UpdateQuoteBillToId(ctx context.Context, id int64, contactId int64) (*Quote, error)
	UpdateQuoteShipToId(ctx context.Context, id int64, contactId int64) (*Quote, error)
	UpdateQuoteSubTotal(ctx context.Context, id int64, subTotal float64) (*Quote, error)
	UpdateQuoteTotal(ctx context.Context, id int64, total float64) (*Quote, error)
	UpdateQuoteBalanceDue(ctx context.Context, id int64, balanceDue float64) (*Quote, error)
	UpdateQuoteBalancePercentDue(ctx context.Context, id int64, balancePercentDue float64) (*Quote, error)
	UpdateQuoteBalanceDueOn(ctx context.Context, id int64, balanceDueOn *time.Time) (*Quote, error)
	UpdateQuotePayUrl(ctx context.Context, id int64, payUrl string) (*Quote, error)
	UpdateQuoteSent(ctx context.Context, id int64, sent bool) (*Quote, error)
	UpdateQuoteSentOn(ctx context.Context, id int64, sentOn *time.Time) (*Quote, error)
	UpdateQuoteSold(ctx context.Context, id int64, sold bool) (*Quote, error)
	UpdateQuoteSoldOn(ctx context.Context, id int64, soldOn *time.Time) (*Quote, error)
	LockQuote(ctx context.Context, id int64) (*Quote, error)
	DeleteQuote(ctx context.Context, id int64) (*Quote, error)

	QuoteAddLineItem(ctx context.Context, id int64, lineItemId int64) (*Quote, error)
	QuoteRemoveLineItem(ctx context.Context, id int64, lineItemId int64) (*Quote, error)
	QuoteAddAdjustment(ctx context.Context, id int64, adjustmentId int64) (*Quote, error)
	QuoteRemoveAdjustment(ctx context.Context, id int64, adjustmentId int64) (*Quote, error)

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

	CreateAdjustment(ctx context.Context, quoteId int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error)
	GetAdjustment(ctx context.Context, id int64) (*Adjustment, error)
	UpdateAdjustment(ctx context.Context, id int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error)
	RemoveAdjustment(ctx context.Context, id int64) error

	// ////////////
	// CONTACT
	// ////////////

	GetContact(ctx context.Context, id int64) (*Contact, error)

	// ////////////
	// HELPER
	// ////////////

	Transaction(func(ctx context.Context) error) error
}
