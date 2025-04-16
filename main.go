package pricey

import (
	"context"

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
	}
}

type PriceyPricebook struct {
	store Store
}

func (v *PriceyPricebook) Create(ctx context.Context, name, description string) (*Pricebook, error) {
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

func (v *PriceyCategory) Create(ctx context.Context, name, description string) (*Category, error) {
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

func (v *PriceyItem) Create(ctx context.Context, categoryId int64, name, description string) (*Item, error) {
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

func (v *PriceyPrice) Create(ctx context.Context, itemId int64, amount float64) (*Price, error) {
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

func (v *PriceyTag) Create(ctx context.Context, pricebookId int64, name, description string) (*Tag, error) {
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

func (v *PriceyImage) Create(ctx context.Context, data []byte) (int64, error) {
	return v.store.CreateImage(ctx, data)
}

// todo: get url, base64, data, delete
// todo: quote stuff
