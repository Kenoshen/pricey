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
	GetPricebook(ctx context.Context, id int) (*Pricebook, error)
	GetPricebooks(ctx context.Context) ([]*Pricebook, error)
	UpdatePricebook(ctx context.Context, pb Pricebook) (*Pricebook, error)
	DeletePricebook(ctx context.Context, id int) error
	RecoverPricebook(ctx context.Context, id int) error

	// ////////////
	// CATEGORY
	// ////////////

	CreateCategory(ctx context.Context, name, description string) (*Category, error)
	GetCategory(ctx context.Context, id int) (*Category, error)
	GetCategories(ctx context.Context, pricebookId int) ([]*Category, error)
	UpdateCategoryInfo(ctx context.Context, id int, name, description string) (*Category, error)
	UpdateCategoryImage(ctx context.Context, id int, imageId, thumbnailId int) (*Category, error)
	MoveCategory(ctx context.Context, id int, parentId int) (*Category, error)
	DeleteCategory(ctx context.Context, id int) error
	DeletePricebookCategories(ctx context.Context, pricebookId int) error
	RecoverCategory(ctx context.Context, id int) error
	RecoverPricebookCategories(ctx context.Context, pricebookId int) error

	// ////////////
	// ITEMS
	// ////////////

	CreateItem(ctx context.Context, categoryId int, name, description string) (*Item, error)
	GetItem(ctx context.Context, id int) (*Item, error)
	GetSimpleItem(ctx context.Context, id int) (*SimpleItem, error)
	GetItemsInCategory(ctx context.Context, categoryId int) ([]*Item, error)
	MoveItem(ctx context.Context, id int, newCategoryId int) (*Item, error)
	UpdateItemInfo(ctx context.Context, id int, code, sku, name, description string) (*Item, error)
	UpdateItemCost(ctx context.Context, id int, cost int) (*Item, error)
	AddItemPrice(ctx context.Context, id int, priceId int) (*Item, error)
	RemoveItemPrice(ctx context.Context, id int, priceId int) (*Item, error)
	AddItemTag(ctx context.Context, id int, tagId int) (*Item, error)
	RemoveItemTag(ctx context.Context, id int, tagId int) (*Item, error)
	RemoveTagFromItems(ctx context.Context, pricebookId, tagId int) error
	UpdateItemHideFromCustomer(ctx context.Context, id int, hideFromCustomer bool) (*Item, error)
	UpdateItemImage(ctx context.Context, id int, imageId, thumbnailId int) (*Item, error)
	SearchItemsInPricebook(ctx context.Context, pricebookId int, search string) ([]*Item, error)
	DeleteItem(ctx context.Context, id int) error
	DeleteCategoryItems(ctx context.Context, categoryId int) error
	DeletePricebookItems(ctx context.Context, pricebookId int) error
	RecoverItem(ctx context.Context, id int) error
	RecoverCategoryItems(ctx context.Context, categoryId int) error
	RecoverPricebookItems(ctx context.Context, pricebookId int) error

	// ////////////
	// SUB ITEM
	// ////////////

	AddSubItem(ctx context.Context, id int, subItemId int, quantity int) (*Item, error)
	UpdateSubItemQuantity(ctx context.Context, id int, subItemId int, quantity int) (*Item, error)
	RemoveSubItem(ctx context.Context, id int, subItemId int) (*Item, error)

	// ////////////
	// PRICE
	// ////////////

	CreatePrice(ctx context.Context, itemId int, amount int) (*Price, error)
	GetPrice(ctx context.Context, id int) (*Price, error)
	GetPricesByItem(ctx context.Context, itemId int) ([]*Price, error)
	MovePricesByItem(ctx context.Context, itemId, categoryId int) error
	UpdatePrice(ctx context.Context, p Price) (*Price, error)
	DeletePrice(ctx context.Context, id int) error
	DeletePricesByItem(ctx context.Context, itemId int) error
	DeleteCategoryPrices(ctx context.Context, categoryId int) error
	DeletePricebookPrices(ctx context.Context, pricebookId int) error
	RecoverPricesByItem(ctx context.Context, itemId int) error
	RecoverCategoryPrices(ctx context.Context, categoryId int) error
	RecoverPricebookPrices(ctx context.Context, pricebookId int) error

	// ////////////
	// TAG
	// ////////////

	CreateTag(ctx context.Context, pricebookId int, name, description string) (*Tag, error)
	GetTag(ctx context.Context, id int) (*Tag, error)
	GetTags(ctx context.Context, pricebookId int) ([]*Tag, error)
	UpdateTagInfo(ctx context.Context, id int, name, description string) (*Tag, error)
	SearchTags(ctx context.Context, pricebookId int, search string) ([]*Tag, error)
	DeleteTag(ctx context.Context, id int) error
	DeletePricebookTags(ctx context.Context, id int) error
	RecoverPricebookTags(ctx context.Context, id int) error

	// ////////////
	// IMAGE
	// ////////////

	CreateImage(ctx context.Context, data []byte) (int, error)
	GetImageUrl(ctx context.Context, id int) (string, error)
	GetImageBase64(ctx context.Context, id int) (string, error)
	GetImageData(ctx context.Context, id int) ([]byte, error)
	DeleteImage(ctx context.Context, id int) error

	// ////////////
	// QUOTE
	// ////////////

	CreateQuote(ctx context.Context) (*Quote, error)
	CreateDuplicateQuote(ctx context.Context, quoteId int) (*Quote, error)
	GetQuote(ctx context.Context, id int) (*Quote, error)
	UpdateQuoteCode(ctx context.Context, id int, code string) (*Quote, error)
	UpdateQuoteOrderNumber(ctx context.Context, id int, orderNumber string) (*Quote, error)
	UpdateQuoteLogoId(ctx context.Context, id int, logoId int) (*Quote, error)
	UpdateQuoteIssueDate(ctx context.Context, id int, issueDate *time.Time) (*Quote, error)
	UpdateQuoteExpirationDate(ctx context.Context, id int, expirationDate *time.Time) (*Quote, error)
	UpdateQuotePaymentTerms(ctx context.Context, id int, paymentTerms string) (*Quote, error)
	UpdateQuoteNotes(ctx context.Context, id int, notes string) (*Quote, error)
	UpdateQuoteSenderId(ctx context.Context, id int, contactId int) (*Quote, error)
	UpdateQuoteBillToId(ctx context.Context, id int, contactId int) (*Quote, error)
	UpdateQuoteShipToId(ctx context.Context, id int, contactId int) (*Quote, error)
	UpdateQuoteSubTotal(ctx context.Context, id int, subTotal int) (*Quote, error)
	UpdateQuoteTotal(ctx context.Context, id int, total int) (*Quote, error)
	UpdateQuoteBalanceDue(ctx context.Context, id int, balanceDue int) (*Quote, error)
	UpdateQuoteBalancePercentDue(ctx context.Context, id int, balancePercentDue int) (*Quote, error)
	UpdateQuoteBalanceDueOn(ctx context.Context, id int, balanceDueOn *time.Time) (*Quote, error)
	UpdateQuotePayUrl(ctx context.Context, id int, payUrl string) (*Quote, error)
	UpdateQuoteSent(ctx context.Context, id int, sent bool) (*Quote, error)
	UpdateQuoteSentOn(ctx context.Context, id int, sentOn *time.Time) (*Quote, error)
	UpdateQuoteSold(ctx context.Context, id int, sold bool) (*Quote, error)
	UpdateQuoteSoldOn(ctx context.Context, id int, soldOn *time.Time) (*Quote, error)
	LockQuote(ctx context.Context, id int) (*Quote, error)
	DeleteQuote(ctx context.Context, id int) (*Quote, error)

	QuoteAddLineItem(ctx context.Context, id int, lineItemId int) (*Quote, error)
	QuoteRemoveLineItem(ctx context.Context, id int, lineItemId int) (*Quote, error)
	QuoteAddAdjustment(ctx context.Context, id int, adjustmentId int) (*Quote, error)
	QuoteRemoveAdjustment(ctx context.Context, id int, adjustmentId int) (*Quote, error)

	// ////////////
	// LINE ITEM
	// ////////////

	CreateLineItem(ctx context.Context, quoteId int, description string, quantity, unitPrice int, amount *int) (*LineItem, error)
	CreateSubLineItem(ctx context.Context, quoteId, parentId int, description string, quantity, unitPrice int, amount *int) (*LineItem, error)
	CreateDuplicateLineItem(ctx context.Context, id int) (*LineItem, error)
	GetLineItem(ctx context.Context, id int) (*LineItem, error)
	MoveLineItem(ctx context.Context, id int, parentId *int, index *int) (*LineItem, error)
	UpdateLineItemImage(ctx context.Context, id int, imageId *int) (*LineItem, error)
	UpdateLineItemDescription(ctx context.Context, id int, description string) (*LineItem, error)
	UpdateLineItemQuantity(ctx context.Context, id int, quantity int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemUnitPrice(ctx context.Context, id int, unitPrice int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemAmount(ctx context.Context, id int, amount *int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemOpen(ctx context.Context, id int, open bool) (*LineItem, error)
	DeleteLineItem(ctx context.Context, id int) error

	// ////////////
	// ADJUSTMENT
	// ////////////

	CreateAdjustment(ctx context.Context, quoteId int, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error)
	GetAdjustment(ctx context.Context, id int) (*Adjustment, error)
	UpdateAdjustment(ctx context.Context, id int, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error)
	RemoveAdjustment(ctx context.Context, id int) error

	// ////////////
	// CONTACT
	// ////////////

	GetContact(ctx context.Context, id int) (*Contact, error)

	// ////////////
	// HELPER
	// ////////////

	Transaction(ctx context.Context, f func(ctx context.Context) error) error
}
