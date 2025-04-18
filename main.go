package pricey

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	gotenberg "github.com/starwalkn/gotenberg-go-client/v8"
)

type Pricey struct {
	store     Store
	pdfClient *gotenberg.Client
	Pricebook *priceyPricebook
	Category  *priceyCategory
	Item      *priceyItem
	Tag       *priceyTag
	Image     *priceyImage
	Quote     *priceyQuote
}

func New(store Store, pdfClient *gotenberg.Client) Pricey {
	return Pricey{
		store:     store,
		pdfClient: pdfClient,
		Pricebook: &priceyPricebook{store},
		Category:  &priceyCategory{store},
		Item:      &priceyItem{store, &priceySubItem{store}},
		Tag:       &priceyTag{store},
		Image:     &priceyImage{store},
		Quote: &priceyQuote{
			store:      store,
			LineItem:   &priceyLineItem{store},
			Adjustment: &priceyAdjustment{store},
			Contact:    &priceyContact{store},
			Print:      newPrinter(store, pdfClient),
		},
	}
}

type priceyPricebook struct {
	store Store
}

func (v *priceyPricebook) New(ctx context.Context, name, description string) (*Pricebook, error) {
	return v.store.CreatePricebook(ctx, name, description)
}

func (v *priceyPricebook) Get(ctx context.Context, id int64) (*Pricebook, error) {
	return v.store.GetPricebook(ctx, id)
}

func (v *priceyPricebook) List(ctx context.Context) ([]*Pricebook, error) {
	return v.store.GetPricebooks(ctx)
}

func (v *priceyPricebook) Set(ctx context.Context, pb Pricebook) (*Pricebook, error) {
	return v.store.UpdatePricebook(ctx, pb)
}

func (v *priceyPricebook) Delete(ctx context.Context, id int64) error {
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

func (v *priceyPricebook) Recover(ctx context.Context, id int64) error {
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

type priceyCategory struct {
	store Store
}

func (v *priceyCategory) New(ctx context.Context, name, description string) (*Category, error) {
	return v.store.CreateCategory(ctx, name, description)
}

func (v *priceyCategory) Get(ctx context.Context, id int64) (*Category, error) {
	return v.store.GetCategory(ctx, id)
}

func (v *priceyCategory) List(ctx context.Context, pricebookId int64) ([]*Category, error) {
	return v.store.GetCategories(ctx, pricebookId)
}

func (v *priceyCategory) SetInfo(ctx context.Context, id int64, name, description string) (*Category, error) {
	return v.store.UpdateCategoryInfo(ctx, id, name, description)
}

func (v *priceyCategory) SetImage(ctx context.Context, id, imageId, thumbnailId int64) (*Category, error) {
	return v.store.UpdateCategoryImage(ctx, id, imageId, thumbnailId)
}

func (v *priceyCategory) Move(ctx context.Context, id, parentId int64) (*Category, error) {
	return v.store.MoveCategory(ctx, id, parentId)
}

func (v *priceyCategory) Delete(ctx context.Context, id int64) error {
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

func (v *priceyCategory) Recover(ctx context.Context, id int64) error {
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

type priceyItem struct {
	store   Store
	SubItem *priceySubItem
}

func (v *priceyItem) New(ctx context.Context, categoryId int64, name, description string) (*Item, error) {
	return v.store.CreateItem(ctx, categoryId, name, description)
}

func (v *priceyItem) Get(ctx context.Context, id int64) (*Item, error) {
	return v.store.GetItem(ctx, id)
}

func (v *priceyItem) GetSimple(ctx context.Context, id int64) (*Item, error) {
	return v.store.GetItem(ctx, id)
}

func (v *priceyItem) Category(ctx context.Context, categoryId int64) ([]*Item, error) {
	return v.store.GetItemsInCategory(ctx, categoryId)
}

func (v *priceyItem) Move(ctx context.Context, id, categoryId int64) (*Item, error) {
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

func (v *priceyItem) SetInfo(ctx context.Context, id int64, code, sku, name, description string) (*Item, error) {
	return v.store.UpdateItemInfo(ctx, id, code, sku, name, description)
}

func (v *priceyItem) SetCost(ctx context.Context, id int64, cost float64) (*Item, error) {
	return v.store.UpdateItemCost(ctx, id, cost)
}

func (v *priceyItem) AddPrice(ctx context.Context, id, priceId int64) (*Item, error) {
	return v.store.AddItemPrice(ctx, id, priceId)
}

func (v *priceyItem) RemovePrice(ctx context.Context, id, priceId int64) (*Item, error) {
	return v.store.RemoveItemPrice(ctx, id, priceId)
}

func (v *priceyItem) AddTag(ctx context.Context, id, tagId int64) (*Item, error) {
	return v.store.AddItemTag(ctx, id, tagId)
}

func (v *priceyItem) RemoveTag(ctx context.Context, id, tagId int64) (*Item, error) {
	return v.store.RemoveItemTag(ctx, id, tagId)
}

func (v *priceyItem) SetHideFromCustomer(ctx context.Context, id int64, hideFromCustomer bool) (*Item, error) {
	return v.store.UpdateItemHideFromCustomer(ctx, id, hideFromCustomer)
}

func (v *priceyItem) SetImage(ctx context.Context, id int64, imageId, thumbnailId int64) (*Item, error) {
	return v.store.UpdateItemImage(ctx, id, imageId, thumbnailId)
}

func (v *priceyItem) Search(ctx context.Context, pricebookId int64, search string) ([]*Item, error) {
	return v.store.SearchItemsInPricebook(ctx, pricebookId, search)
}

func (v *priceyItem) Delete(ctx context.Context, id int64) error {
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

func (v *priceyItem) Recover(ctx context.Context, id int64) error {
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

type priceySubItem struct {
	store Store
}

func (v *priceySubItem) Add(ctx context.Context, id, subItemId, quantity int64) (*Item, error) {
	return v.store.AddSubItem(ctx, id, subItemId, quantity)
}

func (v *priceySubItem) SetQuantity(ctx context.Context, id, subItemId, quantity int64) (*Item, error) {
	return v.store.UpdateSubItemQuantity(ctx, id, subItemId, quantity)
}

func (v *priceySubItem) Delete(ctx context.Context, id, subItemId int64) (*Item, error) {
	return v.store.RemoveSubItem(ctx, id, subItemId)
}

type priceyPrice struct {
	store Store
}

func (v *priceyPrice) New(ctx context.Context, itemId int64, amount float64) (*Price, error) {
	return v.store.CreatePrice(ctx, itemId, amount)
}

func (v *priceyPrice) Get(ctx context.Context, id int64) (*Price, error) {
	return v.store.GetPrice(ctx, id)
}

func (v *priceyPrice) Item(ctx context.Context, itemId int64) ([]*Price, error) {
	return v.store.GetPricesByItem(ctx, itemId)
}

func (v *priceyPrice) Update(ctx context.Context, p Price) (*Price, error) {
	return v.store.UpdatePrice(ctx, p)
}

func (v *priceyPrice) Delete(ctx context.Context, id int64) error {
	return v.store.DeletePrice(ctx, id)
}

type priceyTag struct {
	store Store
}

func (v *priceyTag) New(ctx context.Context, pricebookId int64, name, description string) (*Tag, error) {
	return v.store.CreateTag(ctx, pricebookId, name, description)
}

func (v *priceyTag) Get(ctx context.Context, id int64) (*Tag, error) {
	return v.store.GetTag(ctx, id)
}

func (v *priceyTag) List(ctx context.Context, pricebookId int64) ([]*Tag, error) {
	return v.store.GetTags(ctx, pricebookId)
}

func (v *priceyTag) SetInfo(ctx context.Context, id int64, name, description string) (*Tag, error) {
	return v.store.UpdateTagInfo(ctx, id, name, description)
}

func (v *priceyTag) Search(ctx context.Context, pricebookId int64, search string) ([]*Tag, error) {
	return v.store.SearchTags(ctx, pricebookId, search)
}

func (v *priceyTag) Delete(ctx context.Context, id int64) error {
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

type priceyImage struct {
	store Store
}

func (v *priceyImage) New(ctx context.Context, data []byte) (int64, error) {
	return v.store.CreateImage(ctx, data)
}

func (v *priceyImage) Url(ctx context.Context, id int64) (string, error) {
	return v.store.GetImageUrl(ctx, id)
}

func (v *priceyImage) Base64(ctx context.Context, id int64) (string, error) {
	return v.store.GetImageBase64(ctx, id)
}

func (v *priceyImage) Data(ctx context.Context, id int64) ([]byte, error) {
	return v.store.GetImageData(ctx, id)
}

func (v *priceyImage) Delete(ctx context.Context, id int64) error {
	return v.store.DeleteImage(ctx, id)
}

type priceyQuote struct {
	store      Store
	LineItem   *priceyLineItem
	Adjustment *priceyAdjustment
	Contact    *priceyContact
	Print      *priceyPrint
}

func (v *priceyQuote) New(ctx context.Context) (*Quote, error) {
	return v.store.CreateQuote(ctx)
}

func (v *priceyQuote) Duplicate(ctx context.Context, id int64) (*Quote, error) {
	return v.store.CreateDuplicateQuote(ctx, id)
}

func (v *priceyQuote) Get(ctx context.Context, id int64) (*Quote, error) {
	return v.store.GetQuote(ctx, id)
}

func (v *priceyQuote) SetCode(ctx context.Context, id int64, code string) (*Quote, error) {
	return v.store.UpdateQuoteCode(ctx, id, code)
}

func (v *priceyQuote) SetOrderNumber(ctx context.Context, id int64, orderNumber string) (*Quote, error) {
	return v.store.UpdateQuoteOrderNumber(ctx, id, orderNumber)
}

func (v *priceyQuote) SetLogoId(ctx context.Context, id int64, imageId int64) (*Quote, error) {
	return v.store.UpdateQuoteLogoId(ctx, id, imageId)
}

func (v *priceyQuote) SetIssueDate(ctx context.Context, id int64, issueDate *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteIssueDate(ctx, id, issueDate)
}

func (v *priceyQuote) SetExpirationDate(ctx context.Context, id int64, expirationDate *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteExpirationDate(ctx, id, expirationDate)
}

func (v *priceyQuote) SetPaymentTerms(ctx context.Context, id int64, paymentTerms string) (*Quote, error) {
	return v.store.UpdateQuotePaymentTerms(ctx, id, paymentTerms)
}

func (v *priceyQuote) SetNotes(ctx context.Context, id int64, notes string) (*Quote, error) {
	return v.store.UpdateQuoteNotes(ctx, id, notes)
}

func (v *priceyQuote) SetSenderId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteSenderId(ctx, id, contactId)
}

func (v *priceyQuote) SetBillToId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteBillToId(ctx, id, contactId)
}

func (v *priceyQuote) SetShipToId(ctx context.Context, id int64, contactId int64) (*Quote, error) {
	return v.store.UpdateQuoteShipToId(ctx, id, contactId)
}

func (v *priceyQuote) SetSubTotal(ctx context.Context, id int64, subTotal float64) (*Quote, error) {
	return v.store.UpdateQuoteSubTotal(ctx, id, subTotal)
}

func (v *priceyQuote) SetTotal(ctx context.Context, id int64, total float64) (*Quote, error) {
	return v.store.UpdateQuoteTotal(ctx, id, total)
}

func (v *priceyQuote) SetBalanceDue(ctx context.Context, id int64, balanceDue float64) (*Quote, error) {
	return v.store.UpdateQuoteBalanceDue(ctx, id, balanceDue)
}

func (v *priceyQuote) SetBalancePercentDue(ctx context.Context, id int64, balancePercentDue float64) (*Quote, error) {
	return v.store.UpdateQuoteBalancePercentDue(ctx, id, balancePercentDue)
}

func (v *priceyQuote) SetBalanceDueOn(ctx context.Context, id int64, balanceDueOn *time.Time) (*Quote, error) {
	return v.store.UpdateQuoteBalanceDueOn(ctx, id, balanceDueOn)
}

func (v *priceyQuote) SetPayUrl(ctx context.Context, id int64, payUrl string) (*Quote, error) {
	return v.store.UpdateQuotePayUrl(ctx, id, payUrl)
}

func (v *priceyQuote) SetSent(ctx context.Context, id int64, sent bool) (*Quote, error) {
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

func (v *priceyQuote) SetSold(ctx context.Context, id int64, sold bool) (*Quote, error) {
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

func (v *priceyQuote) Lock(ctx context.Context, id int64) (*Quote, error) {
	return v.store.LockQuote(ctx, id)
}

func (v *priceyQuote) Delete(ctx context.Context, id int64) (*Quote, error) {
	return v.store.DeleteQuote(ctx, id)
}

type priceyLineItem struct {
	store Store
}

func (v *priceyLineItem) New(ctx context.Context, quoteId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error) {
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

func (v *priceyLineItem) NewSub(ctx context.Context, quoteId, parentId int64, description string, quantity, unitPrice float64, amount *float64) (*LineItem, error) {
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

func (v *priceyLineItem) Duplicate(ctx context.Context, id int64) (*LineItem, error) {
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

func (v *priceyLineItem) Get(ctx context.Context, id int64) (*LineItem, error) {
	return v.store.GetLineItem(ctx, id)
}

func (v *priceyLineItem) Move(ctx context.Context, id int64, parentId *int64) (*LineItem, error) {
	return v.store.MoveLineItem(ctx, id, parentId)
}

func (v *priceyLineItem) SetImage(ctx context.Context, id int64, imageId *int64) (*LineItem, error) {
	return v.store.UpdateLineItemImage(ctx, id, imageId)
}

func (v *priceyLineItem) SetDescription(ctx context.Context, id int64, description string) (*LineItem, error) {
	return v.store.UpdateLineItemDescription(ctx, id, description)
}

func (v *priceyLineItem) SetQuantity(ctx context.Context, id int64, quantity float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemQuantity(ctx, id, quantity, prefix, suffix)
}

func (v *priceyLineItem) SetUnitPrice(ctx context.Context, id int64, unitPrice float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemUnitPrice(ctx, id, unitPrice, prefix, suffix)
}

func (v *priceyLineItem) SetAmount(ctx context.Context, id int64, amount *float64, prefix, suffix string) (*LineItem, error) {
	return v.store.UpdateLineItemAmount(ctx, id, amount, prefix, suffix)
}

func (v *priceyLineItem) SetOpen(ctx context.Context, id int64, open bool) (*LineItem, error) {
	return v.store.UpdateLineItemOpen(ctx, id, open)
}

func (v *priceyLineItem) Delete(ctx context.Context, id int64) error {
	return v.store.DeleteLineItem(ctx, id)
}

type priceyAdjustment struct {
	store Store
}

func (v *priceyAdjustment) New(ctx context.Context, quoteId int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error) {
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

func (v *priceyAdjustment) Get(ctx context.Context, id int64) (*Adjustment, error) {
	return v.store.GetAdjustment(ctx, id)
}

func (v *priceyAdjustment) Update(ctx context.Context, id int64, description string, amount float64, adjustmentType AdjustmentType) (*Adjustment, error) {
	return v.store.UpdateAdjustment(ctx, id, description, amount, adjustmentType)
}

func (v *priceyAdjustment) Delete(ctx context.Context, id int64) error {
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

type priceyContact struct {
	store Store
}

func (v *priceyContact) Get(ctx context.Context, id int64) (*Contact, error) {
	return v.store.GetContact(ctx, id)
}

type priceyPrint struct {
	store            Store
	pdfClient        *gotenberg.Client
	standardTemplate *template.Template
}

//go:embed templates/standard.html
var standardTemplate string

func newPrinter(store Store, pdfClient *gotenberg.Client) *priceyPrint {
	funcs := template.FuncMap{
		"pennies":      pennies,
		"quantity":     quantity,
		"depthPadding": depthPadding,
	}
	standardTemplate, err := template.New("standard").Funcs(funcs).Parse(standardTemplate)
	if err != nil {
		panic("failed to parse standard template: " + err.Error())
	}

	return &priceyPrint{
		store:            store,
		pdfClient:        pdfClient,
		standardTemplate: standardTemplate,
	}
}

func pennies(v int64) string {
	if v == 0 {
		return ""
	}
	d := fmt.Sprintf("%d", v/100)
	ld := len(d)
	sb := strings.Builder{}
	for i := ld - 1; i >= 0; i-- {
		c := d[i]
		if ld-i > 1 && (ld-i)%3 == 1 {
			sb.WriteString(",")
		}
		sb.WriteByte(c)
	}
	d = sb.String()
	ld = len(d)
	sb = strings.Builder{}
	for i := ld - 1; i >= 0; i-- {
		sb.WriteByte(d[i])
	}

	pens := v % 100
	return fmt.Sprintf("%s.%02d", sb.String(), pens)
}

func quantity(v int64) string {
	if v == 0 {
		return ""
	}
	whole := v / 100
	frac := v % 100
	if frac == 0 {
		return fmt.Sprintf("%d", whole)
	} else if frac%10 == 0 {
		return fmt.Sprintf("%d.%d", whole, frac/10)
	}
	return fmt.Sprintf("%d.%d", whole, frac)
}

func depthPadding(v int, paddingPixels, base int) int {
	return (v * paddingPixels) + base
}

func (v *priceyPrint) GetFullQuote(ctx context.Context, id int64) (*FullQuote, error) {
	q := &FullQuote{Id: id}
	return q, v.store.Transaction(func(ctx context.Context) error {
		quote, err := v.store.GetQuote(ctx, id)
		if err != nil {
			return err
		}
		if quote.LogoId >= 0 {
			imageUrl, err := v.store.GetImageUrl(ctx, quote.LogoId)
			if err != nil {
				return err
			}
			if imageUrl != "" {
				q.Logo = &Image{
					Id:  quote.LogoId,
					Url: imageUrl,
				}
			}
		}
		if quote.SenderId >= 0 {
			q.Sender, err = v.store.GetContact(ctx, quote.SenderId)
			if err != nil {
				return err
			}
		}
		if quote.BillToId >= 0 {
			q.BillTo, err = v.store.GetContact(ctx, quote.BillToId)
			if err != nil {
				return err
			}
		}
		if quote.ShipToId >= 0 {
			q.ShipTo, err = v.store.GetContact(ctx, quote.ShipToId)
			if err != nil {
				return err
			}
		}
		// todo: get line items
		// todo: calculate subTotal
		// todo: get adjustments
		// todo: calculate total
		// todo: calculate balanceDue

		return nil
	})
}

func (v *priceyPrint) Standard(ctx context.Context, id int64, w io.Writer) error {
	q, err := v.GetFullQuote(ctx, id)
	if err != nil {
		return err
	}
	buf := bytes.Buffer{}
	err = v.standardTemplate.Execute(&buf, q)
	if err != nil {
		return err
	}
	resp, err := print(v.pdfClient, &buf)
	if err != nil {
		return err
	}
	defer resp.Close()
	_, err = io.Copy(w, resp)
	if err != nil {
		return err
	}
	return nil
}

func (v *priceyPrint) StandardHTML(ctx context.Context, id int64, w io.Writer) error {
	q, err := v.GetFullQuote(ctx, id)
	if err != nil {
		return err
	}
	return v.standardTemplate.Execute(w, q)
}
