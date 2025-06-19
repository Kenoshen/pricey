package pricey

import (
	"context"
	"errors"
	"time"
)

type Store interface {
	// ////////////
	// PRICEBOOK
	// ////////////

	CreatePricebook(ctx context.Context, name, description string) (*Pricebook, error)
	GetPricebook(ctx context.Context, id ID) (*Pricebook, error)
	GetPricebooks(ctx context.Context) ([]*Pricebook, error)
	UpdatePricebook(ctx context.Context, pb Pricebook) (*Pricebook, error)
	DeletePricebook(ctx context.Context, id ID) error
	RecoverPricebook(ctx context.Context, id ID) error

	// ////////////
	// CATEGORY
	// ////////////

	CreateCategory(ctx context.Context, pricebookId ID, name, description string) (*Category, error)
	GetCategory(ctx context.Context, id ID) (*Category, error)
	GetCategories(ctx context.Context, pricebookId ID) ([]*Category, error)
	UpdateCategoryInfo(ctx context.Context, id ID, name, description string) (*Category, error)
	UpdateCategoryImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Category, error)
	MoveCategory(ctx context.Context, id ID, parentId ID) (*Category, error)
	DeleteCategory(ctx context.Context, id ID) error
	DeletePricebookCategories(ctx context.Context, pricebookId ID) error
	RecoverCategory(ctx context.Context, id ID) error
	RecoverPricebookCategories(ctx context.Context, pricebookId ID) error

	// ////////////
	// ITEMS
	// ////////////

	CreateItem(ctx context.Context, categoryId ID, name, description string) (*Item, error)
	GetItem(ctx context.Context, id ID) (*Item, error)
	GetSimpleItem(ctx context.Context, id ID) (*SimpleItem, error)
	GetItemsInCategory(ctx context.Context, categoryId ID) ([]*Item, error)
	MoveItem(ctx context.Context, id ID, newCategoryId ID) (*Item, error)
	UpdateItemInfo(ctx context.Context, id ID, code, sku, name, description string) (*Item, error)
	UpdateItemCost(ctx context.Context, id ID, cost int) (*Item, error)
	AddItemPrice(ctx context.Context, id ID, priceId ID) (*Item, error)
	RemoveItemPrice(ctx context.Context, id ID, priceId ID) (*Item, error)
	AddItemTag(ctx context.Context, id ID, tagId ID) (*Item, error)
	RemoveItemTag(ctx context.Context, id ID, tagId ID) (*Item, error)
	RemoveTagFromItems(ctx context.Context, pricebookId, tagId ID) error
	UpdateItemHideFromCustomer(ctx context.Context, id ID, hideFromCustomer bool) (*Item, error)
	UpdateItemImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Item, error)
	SearchItemsInPricebook(ctx context.Context, pricebookId ID, search string) ([]*Item, error)
	DeleteItem(ctx context.Context, id ID) error
	DeleteCategoryItems(ctx context.Context, categoryId ID) error
	DeletePricebookItems(ctx context.Context, pricebookId ID) error
	RecoverItem(ctx context.Context, id ID) error
	RecoverCategoryItems(ctx context.Context, categoryId ID) error
	RecoverPricebookItems(ctx context.Context, pricebookId ID) error

	// ////////////
	// SUB ITEM
	// ////////////

	AddSubItem(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error)
	UpdateSubItemQuantity(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error)
	RemoveSubItem(ctx context.Context, id ID, subItemId ID) (*Item, error)

	// ////////////
	// PRICE
	// ////////////

	CreatePrice(ctx context.Context, itemId ID, amount int) (*Price, error)
	GetPrice(ctx context.Context, id ID) (*Price, error)
	GetPricesByItem(ctx context.Context, itemId ID) ([]*Price, error)
	MovePricesByItem(ctx context.Context, itemId, categoryId ID) error
	UpdatePrice(ctx context.Context, p Price) (*Price, error)
	DeletePrice(ctx context.Context, id ID) error
	DeletePricesByItem(ctx context.Context, itemId ID) error
	DeleteCategoryPrices(ctx context.Context, categoryId ID) error
	DeletePricebookPrices(ctx context.Context, pricebookId ID) error
	RecoverPricesByItem(ctx context.Context, itemId ID) error
	RecoverCategoryPrices(ctx context.Context, categoryId ID) error
	RecoverPricebookPrices(ctx context.Context, pricebookId ID) error

	// ////////////
	// TAG
	// ////////////

	CreateTag(ctx context.Context, pricebookId ID, name, description string) (*Tag, error)
	GetTag(ctx context.Context, id ID) (*Tag, error)
	GetTags(ctx context.Context, pricebookId ID) ([]*Tag, error)
	UpdateTagInfo(ctx context.Context, id ID, name, description string) (*Tag, error)
	SearchTags(ctx context.Context, pricebookId ID, search string) ([]*Tag, error)
	DeleteTag(ctx context.Context, id ID) error
	DeletePricebookTags(ctx context.Context, id ID) error
	RecoverPricebookTags(ctx context.Context, id ID) error

	// ////////////
	// IMAGE
	// ////////////

	CreateImage(ctx context.Context, data []byte) (ID, error)
	GetImageUrl(ctx context.Context, id ID) (string, error)
	GetImageBase64(ctx context.Context, id ID) (string, error)
	GetImageData(ctx context.Context, id ID) ([]byte, error)
	DeleteImage(ctx context.Context, id ID) error

	// ////////////
	// QUOTE
	// ////////////

	CreateQuote(ctx context.Context) (*Quote, error)
	CreateDuplicateQuote(ctx context.Context, quoteId ID) (*Quote, error)
	GetQuote(ctx context.Context, id ID) (*Quote, error)
	UpdateQuoteCode(ctx context.Context, id ID, code string) (*Quote, error)
	UpdateQuoteOrderNumber(ctx context.Context, id ID, orderNumber string) (*Quote, error)
	UpdateQuoteLogoId(ctx context.Context, id ID, logoId ID) (*Quote, error)
	UpdateQuoteIssueDate(ctx context.Context, id ID, issueDate *time.Time) (*Quote, error)
	UpdateQuoteExpirationDate(ctx context.Context, id ID, expirationDate *time.Time) (*Quote, error)
	UpdateQuotePaymentTerms(ctx context.Context, id ID, paymentTerms string) (*Quote, error)
	UpdateQuoteNotes(ctx context.Context, id ID, notes string) (*Quote, error)
	UpdateQuoteSenderId(ctx context.Context, id ID, contactId ID) (*Quote, error)
	UpdateQuoteBillToId(ctx context.Context, id ID, contactId ID) (*Quote, error)
	UpdateQuoteShipToId(ctx context.Context, id ID, contactId ID) (*Quote, error)
	UpdateQuoteSubTotal(ctx context.Context, id ID, subTotal int) (*Quote, error)
	UpdateQuoteTotal(ctx context.Context, id ID, total int) (*Quote, error)
	UpdateQuoteBalanceDue(ctx context.Context, id ID, balanceDue int) (*Quote, error)
	UpdateQuoteBalancePercentDue(ctx context.Context, id ID, balancePercentDue int) (*Quote, error)
	UpdateQuoteBalanceDueOn(ctx context.Context, id ID, balanceDueOn *time.Time) (*Quote, error)
	UpdateQuotePayUrl(ctx context.Context, id ID, payUrl string) (*Quote, error)
	UpdateQuoteSent(ctx context.Context, id ID, sent bool) (*Quote, error)
	UpdateQuoteSentOn(ctx context.Context, id ID, sentOn *time.Time) (*Quote, error)
	UpdateQuoteSold(ctx context.Context, id ID, sold bool) (*Quote, error)
	UpdateQuoteSoldOn(ctx context.Context, id ID, soldOn *time.Time) (*Quote, error)
	LockQuote(ctx context.Context, id ID) (*Quote, error)
	DeleteQuote(ctx context.Context, id ID) (*Quote, error)

	QuoteAddLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error)
	QuoteRemoveLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error)
	QuoteAddAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error)
	QuoteRemoveAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error)

	// ////////////
	// LINE ITEM
	// ////////////

	CreateLineItem(ctx context.Context, quoteId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error)
	CreateSubLineItem(ctx context.Context, quoteId, parentId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error)
	CreateDuplicateLineItem(ctx context.Context, id ID) (*LineItem, error)
	GetLineItem(ctx context.Context, id ID) (*LineItem, error)
	MoveLineItem(ctx context.Context, id ID, parentId *ID, index *int) (*LineItem, error)
	UpdateLineItemImage(ctx context.Context, id ID, imageId *ID) (*LineItem, error)
	UpdateLineItemDescription(ctx context.Context, id ID, description string) (*LineItem, error)
	UpdateLineItemQuantity(ctx context.Context, id ID, quantity int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemUnitPrice(ctx context.Context, id ID, unitPrice int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemAmount(ctx context.Context, id ID, amount *int, prefix, suffix string) (*LineItem, error)
	UpdateLineItemOpen(ctx context.Context, id ID, open bool) (*LineItem, error)
	DeleteLineItem(ctx context.Context, id ID) error

	// ////////////
	// ADJUSTMENT
	// ////////////

	CreateAdjustment(ctx context.Context, quoteId ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error)
	GetAdjustment(ctx context.Context, id ID) (*Adjustment, error)
	UpdateAdjustment(ctx context.Context, id ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error)
	RemoveAdjustment(ctx context.Context, id ID) error

	// ////////////
	// CONTACT
	// ////////////

	GetContact(ctx context.Context, id ID) (*Contact, error)

	// ////////////
	// HELPER
	// ////////////

	Transaction(ctx context.Context, f func(ctx context.Context) error) error
}

type Auth interface {
	CreateToken(orgId, groupId, userId ID, claims map[string]interface{}) (string, error)
}

type OrgGroupExtractor = func(ctx context.Context) (orgId ID, groupId ID, err error)

var (
	EmptyOrgIdInContextError   = errors.New("orgId is empty in context")
	EmptyGroupIdInContextError = errors.New("groupId is empty in context")
)

func OrgGroupExtractorConfig(orgKey any, groupKey any) OrgGroupExtractor {
	return func(ctx context.Context) (orgId ID, groupId ID, err error) {
		orgId = ctx.Value(orgKey).(string)
		if orgId == "" {
			err = EmptyOrgIdInContextError
			return
		}
		groupId = ctx.Value(groupKey).(string)
		if groupId == "" {
			err = EmptyGroupIdInContextError
			return
		}
		return
	}
}
