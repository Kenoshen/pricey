package pricey

import (
	"context"
	"time"

	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
)

type Pricey struct {
	store     Store
	pdfClient *gotenberg.Client
	Pricebook *PriceyPricebook
	Category  *PriceyCategory
	Item      *PriceyItem
	Tag       *PriceyTag
	Image     *PriceyImage
	Quote     *PriceyQuote
}

func New(store Store, pdfClient *gotenberg.Client) Pricey {
	return Pricey{
		store:     store,
		pdfClient: pdfClient,
		Pricebook: &PriceyPricebook{store},
		Category:  &PriceyCategory{store},
		Item:      &PriceyItem{store, &PriceySubItem{store}},
		Tag:       &PriceyTag{store},
		Image:     &PriceyImage{store},
		Quote:     &PriceyQuote{store, pdfClient, &PriceyLineItem{store}, &PriceyAdjustment{store}, &PriceyContact{store}},
	}
}

type PriceyPricebook struct {
	store Store
}

func (v *PriceyPricebook) New(ctx context.Context, name, description string) (*Pricebook, error) {
	return v.store.CreatePricebook(ctx, name, description)
}

func (v *PriceyPricebook) Get(ctx context.Context, id int64) (*Pricebook, error) {
	return v.store.GetPricebook(ctx, id)
}

func (v *PriceyPricebook) List(ctx context.Context) ([]*Pricebook, error) {
	return v.store.GetPricebooks(ctx)
}

func (v *PriceyPricebook) Set(ctx context.Context, pb Pricebook) (*Pricebook, error) {
	return v.store.UpdatePricebook(ctx, pb)
}

func (v *PriceyPricebook) Delete(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.DeletePricebook(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeletePricebookCategories(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeletePricebookItems(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeletePricebookPrices(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeletePricebookTags(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyPricebook) Recover(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.RecoverPricebook(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverPricebookCategories(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverPricebookItems(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverPricebookPrices(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverPricebookTags(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

type PriceyCategory struct {
	store Store
}

func (v *PriceyCategory) New(ctx context.Context, name, description string) (*Category, error) {
	return v.store.CreateCategory(ctx, name, description)
}

func (v *PriceyCategory) Get(ctx context.Context, id int64) (*Category, error) {
	return v.store.GetCategory(ctx, id)
}

func (v *PriceyCategory) List(ctx context.Context, pricebookId int64) ([]*Category, error) {
	return v.store.GetCategories(ctx, pricebookId)
}

func (v *PriceyCategory) SetInfo(ctx context.Context, id int64, name, description string) (*Category, error) {
	return v.store.UpdateCategoryInfo(ctx, id, name, description)
}

func (v *PriceyCategory) SetImage(ctx context.Context, id, imageId, thumbnailId int64) (*Category, error) {
	return v.store.UpdateCategoryImage(ctx, id, imageId, thumbnailId)
}

func (v *PriceyCategory) Move(ctx context.Context, id, parentId int64) (*Category, error) {
	return v.store.MoveCategory(ctx, id, parentId)
}

func (v *PriceyCategory) Delete(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.DeleteCategory(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeleteCategoryItems(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.DeleteCategoryPrices(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyCategory) Recover(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.RecoverCategory(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverCategoryItems(ctx, id)
		if err != nil {
			return err
		}
		err = v.store.RecoverCategoryPrices(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

type PriceyItem struct {
	store   Store
	SubItem *PriceySubItem
}

func (v *PriceyItem) New(ctx context.Context, categoryId int64, name, description string) (*Item, error) {
	return v.store.CreateItem(ctx, categoryId, name, description)
}

func (v *PriceyItem) Get(ctx context.Context, id int64) (*Item, error) {
	return v.store.GetItem(ctx, id)
}

func (v *PriceyItem) GetSimple(ctx context.Context, id int64) (*Item, error) {
	return v.store.GetItem(ctx, id)
}

func (v *PriceyItem) Category(ctx context.Context, categoryId int64) ([]*Item, error) {
	return v.store.GetItemsInCategory(ctx, categoryId)
}

func (v *PriceyItem) Move(ctx context.Context, id, categoryId int64) (*Item, error) {
	var item *Item
	return item, v.store.Transaction(func(ctx context.Context) error {
		var err error
		item, err = v.store.MoveItem(ctx, id, categoryId)
		if err != nil {
			return err
		}

		err = v.store.MovePricesByItem(ctx, id, categoryId)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyItem) SetInfo(ctx context.Context, id int64, code, sku, name, description string) (*Item, error) {
	return v.store.UpdateItemInfo(ctx, id, code, sku, name, description)
}

func (v *PriceyItem) SetCost(ctx context.Context, id int64, cost float64) (*Item, error) {
	return v.store.UpdateItemCost(ctx, id, cost)
}

func (v *PriceyItem) AddPrice(ctx context.Context, id, priceId int64) (*Item, error) {
	return v.store.AddItemPrice(ctx, id, priceId)
}

func (v *PriceyItem) RemovePrice(ctx context.Context, id, priceId int64) (*Item, error) {
	return v.store.RemoveItemPrice(ctx, id, priceId)
}

func (v *PriceyItem) AddTag(ctx context.Context, id, tagId int64) (*Item, error) {
	return v.store.AddItemTag(ctx, id, tagId)
}

func (v *PriceyItem) RemoveTag(ctx context.Context, id, tagId int64) (*Item, error) {
	return v.store.RemoveItemTag(ctx, id, tagId)
}

func (v *PriceyItem) SetHideFromCustomer(ctx context.Context, id int64, hideFromCustomer bool) (*Item, error) {
	return v.store.UpdateItemHideFromCustomer(ctx, id, hideFromCustomer)
}

func (v *PriceyItem) SetImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Item, error) {
	return v.store.UpdateItemImage(ctx, id, imageId, thumbnailId)
}

func (v *PriceyItem) Search(ctx context.Context, pricebookId int64, search string) ([]*Item, error) {
	return v.store.SearchItemsInPricebook(ctx, pricebookId, search)
}

func (v *PriceyItem) Delete(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.DeleteItem(ctx, id)
		if err != nil {
			return err
		}

		err = v.store.DeletePricesByItem(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyItem) Recover(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		err := v.store.RecoverItem(ctx, id)
		if err != nil {
			return err
		}

		err = v.store.RecoverPricesByItem(ctx, id)
		if err != nil {
			return err
		}

		return nil
	})
}

type PriceySubItem struct {
	store Store
}

func (v *PriceySubItem) Add(ctx context.Context, id, subItemId, quantity int64) (*Item, error) {
	return v.store.AddSubItem(ctx, id, subItemId, quantity)
}

func (v *PriceySubItem) SetQuantity(ctx context.Context, id, subItemId, quantity int64) (*Item, error) {
	return v.store.UpdateSubItemQuantity(ctx, id, subItemId, quantity)
}

func (v *PriceySubItem) Delete(ctx context.Context, id, subItemId int64) (*Item, error) {
	return v.store.RemoveSubItem(ctx, id, subItemId)
}

type PriceyPrice struct {
	store Store
}

func (v *PriceyPrice) New(ctx context.Context, itemId int64, amount float64) (*Price, error) {
	return v.store.CreatePrice(ctx, itemId, amount)
}

func (v *PriceyPrice) Get(ctx context.Context, id int64) (*Price, error) {
	return v.store.GetPrice(ctx, id)
}

func (v *PriceyPrice) Item(ctx context.Context, itemId int64) ([]*Price, error) {
	return v.store.GetPricesByItem(ctx, itemId)
}

func (v *PriceyPrice) Update(ctx context.Context, p Price) (*Price, error) {
	return v.store.UpdatePrice(ctx, p)
}

func (v *PriceyPrice) Delete(ctx context.Context, id int64) error {
	return v.store.DeletePrice(ctx, id)
}

type PriceyTag struct {
	store Store
}

func (v *PriceyTag) New(ctx context.Context, pricebookId int64, name, description string) (*Tag, error) {
	return v.store.CreateTag(ctx, pricebookId, name, description)
}

func (v *PriceyTag) Get(ctx context.Context, id int64) (*Tag, error) {
	return v.store.GetTag(ctx, id)
}

func (v *PriceyTag) List(ctx context.Context, pricebookId int64) ([]*Tag, error) {
	return v.store.GetTags(ctx, pricebookId)
}

func (v *PriceyTag) SetInfo(ctx context.Context, id int64, name, description string) (*Tag, error) {
	return v.store.UpdateTagInfo(ctx, id, name, description)
}

func (v *PriceyTag) Search(ctx context.Context, pricebookId int64, search string) ([]*Tag, error) {
	return v.store.SearchTags(ctx, pricebookId, search)
}

func (v *PriceyTag) Delete(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		tag, err := v.store.GetTag(ctx, id)
		if err != nil {
			return err
		}
		if tag == nil {
			return nil
		}

		err = v.store.DeleteTag(ctx, id)
		if err != nil {
			return err
		}

		err = v.store.RemoveTagFromItems(ctx, tag.PricebookId, id)
		if err != nil {
			return err
		}

		return nil
	})
}

type PriceyImage struct {
	store Store
}

func (v *PriceyImage) New(ctx context.Context, data []byte) (int64, error) {
	return v.store.CreateImage(ctx, data)
}

func (v *PriceyImage) Url(ctx context.Context, id int64) (string, error) {
	return v.store.GetImageUrl(ctx, id)
}

func (v *PriceyImage) Base64(ctx context.Context, id int64) (string, error) {
	return v.store.GetImageBase64(ctx, id)
}

func (v *PriceyImage) Data(ctx context.Context, id int64) ([]byte, error) {
	return v.store.GetImageData(ctx, id)
}

func (v *PriceyImage) Delete(ctx context.Context, id int64) error {
	return v.store.DeleteImage(ctx, id)
}

type PriceyQuote struct {
	store      Store
	pdfClient  *gotenberg.Client
	LineItem   *PriceyLineItem
	Adjustment *PriceyAdjustment
	Contact    *PriceyContact
}

func (v *PriceyQuote) New(ctx context.Context) (*Quote, error) {
	return v.store.CreateQuote(ctx)
}

func (v *PriceyQuote) FromTemplate(ctx context.Context, templateId int64) (*Quote, error) {
	return v.store.CreateQuoteFromTemplate(ctx, templateId)
}

func (v *PriceyQuote) Duplicate(ctx context.Context, id int64) (*Quote, error) {
	return v.store.CreateDuplicateQuote(ctx, id)
}

func (v *PriceyQuote) Get(ctx context.Context, id int64) (*Quote, error) {
	return v.store.GetQuote(ctx, id)
}

func (v *PriceyQuote) SetCode(ctx context.Context, id int64, code string) (*Quote, error) {
	return v.store.UpdateQuoteCode(ctx, id, code)
}

func (v *PriceyQuote) SetOrderNumber(ctx context.Context, id int64, orderNumber string) (*Quote, error) {
	return v.store.UpdateQuoteOrderNumber(ctx, id, orderNumber)
}

func (v *PriceyQuote) SetLogoId(ctx context.Context, id int64, imageId int64) (*Quote, error) {
	return v.store.UpdateQuoteLogoId(ctx, id, imageId)
}

func (v *PriceyQuote) SetIssueDate(ctx context.Context, id int64, issueDate *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteIssueDate(ctx, id, issueDate)
}

func (v *PriceyQuote) SetExpirationDate(ctx context.Context, id int64, expirationDate *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteExpirationDate(ctx, id, expirationDate)
}

func (v *PriceyQuote) SetPaymentTerms(ctx context.Context, id int64, paymentTerms string) (*Quote, error) {
	return v.store.UpdateQuotePaymentTerms(ctx, id, paymentTerms)
}

func (v *PriceyQuote) SetNotes(ctx context.Context, id int64, notes string) (*Quote, error) {
	return v.store.UpdateQuoteNotes(ctx, id, notes)
}

func (v *PriceyQuote) SetSenderId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteSenderId(ctx, id, contactId)
}

func (v *PriceyQuote) SetBillToId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteBillToId(ctx, id, contactId)
}

func (v *PriceyQuote) SetShipToId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteShipToId(ctx, id, contactId)
}

func (v *PriceyQuote) SetSubTotal(ctx context.Context, id int64, subTotal float64) (*Quote, error) {
	return v.store.UpdateQuoteSubTotal(ctx, id, subTotal)
}

func (v *PriceyQuote) SetTotal(ctx context.Context, id int64, total float64) (*Quote, error) {
	return v.store.UpdateQuoteTotal(ctx, id, total)
}

func (v *PriceyQuote) SetBalanceDue(ctx context.Context, id int64, balanceDue float64) (*Quote, error) {
	return v.store.UpdateQuoteBalanceDue(ctx, id, balanceDue)
}

func (v *PriceyQuote) SetBalancePercentDue(ctx context.Context, id int64, balancePercentDue float64) (*Quote, error) {
	return v.store.UpdateQuoteBalancePercentDue(ctx, id, balancePercentDue)
}

func (v *PriceyQuote) SetBalanceDueOn(ctx context.Context, id int64, balanceDueOn *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteBalanceDueOn(ctx, id, balanceDueOn)
}

func (v *PriceyQuote) SetPayUrl(ctx context.Context, id int64, payUrl string) (*Quote, error) {
	return v.store.UpdateQuotePayUrl(ctx, id, payUrl)
}

func (v *PriceyQuote) SetSent(ctx context.Context, id int64, sent bool) (*Quote, error) {
	var q *Quote
	return q, v.store.Transaction(func(ctx context.Context) error {
		var err error
		q, err = v.store.UpdateQuoteSent(ctx, id, sent)
		if err != nil {
			return err
		}
		if sent {
			now := time.Now()
			q, err = v.store.UpdateQuoteSentOn(ctx, id, &now)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (v *PriceyQuote) SetSold(ctx context.Context, id int64, sold bool) (*Quote, error) {
	var q *Quote
	return q, v.store.Transaction(func(ctx context.Context) error {
		var err error
		q, err = v.store.UpdateQuoteSold(ctx, id, sold)
		if err != nil {
			return err
		}
		if sold {
			now := time.Now()
			q, err = v.store.UpdateQuoteSoldOn(ctx, id, &now)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (v *PriceyQuote) Lock(ctx context.Context, id int64) (*Quote, error) {
	return v.store.LockQuote(ctx, id)
}

func (v *PriceyQuote) Delete(ctx context.Context, id int64) (*Quote, error) {
	return v.store.DeleteQuote(ctx, id)
}

type PriceyLineItem struct {
	store Store
}

func (v *PriceyLineItem) New(ctx context.Context, quoteId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error) {
	var item *LineItem
	return item, v.store.Transaction(func(ctx context.Context) error {
		var err error
		item, err = v.store.CreateLineItem(ctx, quoteId, description, quantity, unitPrice, amount)
		if err != nil {
			return err
		}

		// TODO: need to figure out how the line items are ordered in the list
		_, err = v.store.QuoteAddLineItem(ctx, quoteId, item.Id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyLineItem) NewSub(ctx context.Context, quoteId, parentId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error) {
	var item *LineItem
	return item, v.store.Transaction(func(ctx context.Context) error {
		var err error
		item, err = v.store.CreateSubLineItem(ctx, quoteId, parentId, description, quantity, unitPrice, amount)
		if err != nil {
			return err
		}

		// TODO: need to figure out how the line items are ordered in the list
		_, err = v.store.QuoteAddLineItem(ctx, item.QuoteId, item.Id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyLineItem) Duplicate(ctx context.Context, id int64) (*LineItem, error) {
	var item *LineItem
	return item, v.store.Transaction(func(ctx context.Context) error {
		var err error
		item, err = v.store.CreateDuplicateLineItem(ctx, id)
		if err != nil {
			return err
		}

		// TODO: need to figure out how the line items are ordered in the list
		_, err = v.store.QuoteAddLineItem(ctx, item.QuoteId, item.Id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyLineItem) Get(ctx context.Context, id int64) (*LineItem, error) {
	return v.store.GetLineItem(ctx, id)
}

func (v *PriceyLineItem) Move(ctx context.Context, id int64, parentId *int64) (*LineItem, error) {
	return v.store.MoveLineItem(ctx, id, parentId)
}

func (v *PriceyLineItem) SetImage(ctx context.Context, id int64, imageId *int64) (*LineItem, error) {
	return v.store.UpdateLineItemImage(ctx, id, imageId)
}

func (v *PriceyLineItem) SetDescription(ctx context.Context, id int64, description string) (*LineItem, error) {
	return v.store.UpdateLineItemDescription(ctx, id, description)
}

func (v *PriceyLineItem) SetQuantity(ctx context.Context, id int64, quantity float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemQuantity(ctx, id, quantity, prefix, suffix)
}

func (v *PriceyLineItem) SetUnitPrice(ctx context.Context, id int64, unitPrice float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemUnitPrice(ctx, id, unitPrice, prefix, suffix)
}

func (v *PriceyLineItem) SetAmount(ctx context.Context, id int64, amount *float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemAmount(ctx, id, amount, prefix, suffix)
}

func (v *PriceyLineItem) SetOpen(ctx context.Context, id int64, open bool) (*LineItem, error) {
	return v.store.UpdateLineItemOpen(ctx, id, open)
}

func (v *PriceyLineItem) Delete(ctx context.Context, id int64) error {
	return v.store.DeleteLineItem(ctx, id)
}

type PriceyAdjustment struct {
	store Store
}

func (v *PriceyAdjustment) New(ctx context.Context, quoteId int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error) {
	var a *Adjustment
	return a, v.store.Transaction(func(ctx context.Context) error {
		var err error
		a, err = v.store.CreateAdjustment(ctx, quoteId, description, amount, adjustmentType)
		if err != nil {
			return err
		}

		_, err = v.store.QuoteAddAdjustment(ctx, quoteId, a.Id)
		if err != nil {
			return err
		}

		return nil
	})
}

func (v *PriceyAdjustment) Get(ctx context.Context, id int64) (*Adjustment, error) {
	return v.store.GetAdjustment(ctx, id)
}

func (v *PriceyAdjustment) Update(ctx context.Context, id int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error) {
	return v.store.UpdateAdjustment(ctx, id, description, amount, adjustmentType)
}

func (v *PriceyAdjustment) Delete(ctx context.Context, id int64) error {
	return v.store.Transaction(func(ctx context.Context) error {
		a, err := v.store.GetAdjustment(ctx, id)
		if err != nil {
			return err
		}
		if a == nil {
			return nil
		}

		err = v.store.RemoveAdjustment(ctx, id)
		if err != nil {
			return err
		}

		_, err = v.store.QuoteRemoveAdjustment(ctx, a.QuoteId, id)
		if err != nil {
			return err
		}

		return nil
	})
}

type PriceyContact struct {
	store Store
}

func (v *PriceyContact) Get(ctx context.Context, id int64) (*Contact, error) {
	return v.store.GetContact(ctx, id)
}
