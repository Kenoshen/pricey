package pricey

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Firebase struct {
	app  *firebase.App
	auth *auth.Client
	fire *firestore.Client
	ext  OrgGroupExtractor
}

const (
	PricebookCollection = "pricebook"
	CategoryCollection  = "category"
	ItemCollection      = "item"
	TagCollection       = "tag"
	PriceCollection     = "price"
)

var (
	UnauthorizedOrgError   = errors.New("orgId did not match orgId of document")
	UnauthorizedGroupError = errors.New("groupId did not match groupId of document")
	InvalidPriceIdError    = errors.New("price id is invalid because it does not exist on the sub item")
)

func NewFirebase(ctx context.Context, ext OrgGroupExtractor, config *firebase.Config, opts ...option.ClientOption) (*Firebase, error) {
	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, err
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	fire, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	return &Firebase{
		app:  app,
		auth: auth,
		fire: fire,
		ext:  ext,
	}, nil
}

// CreateToken generates a custom auth token for
func (f *Firebase) CreateToken(ctx context.Context, orgId, groupId, userId ID, claims map[string]any) (string, error) {
	actualClaims := map[string]any{}
	maps.Copy(actualClaims, claims)
	actualClaims["o"] = orgId
	actualClaims["g"] = groupId
	return f.auth.CustomTokenWithClaims(ctx, userId, actualClaims)
}

// get is a wrapper around the firebase collection, doc, get methods and it
// does some checks for values on the data to make sure that org and group ids
// are present and match the context values
func (f *Firebase) get(ctx context.Context, collection, id string, dataOut any) error {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return err
	}
	result, err := f.fire.Collection(PricebookCollection).Doc(id).Get(ctx)
	if err != nil {
		return err
	}
	err = result.DataTo(dataOut)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(dataOut).Elem()
	if v.FieldByName("OrgId").String() != orgId {
		return UnauthorizedOrgError
	}
	if v.FieldByName("GroupId").String() != groupId {
		return UnauthorizedGroupError
	}
	return nil
}

// query wraps the firebase query function iterates through the results and returns
// a slice of snapshots that can be converted to a type using docsToType
func (f *Firebase) query(ctx context.Context, collection string, filters ...firestore.EntityFilter) ([]*firestore.DocumentSnapshot, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}
	finalFilters := []firestore.EntityFilter{
		firestore.PropertyFilter{
			Path:     "orgId",
			Operator: "==",
			Value:    orgId,
		},
		firestore.PropertyFilter{
			Path:     "groupId",
			Operator: "==",
			Value:    groupId,
		},
	}
	finalFilters = append(finalFilters, filters...)
	docs := f.fire.Collection(collection).Query.WhereEntity(firestore.AndFilter{Filters: finalFilters}).Documents(ctx)
	var results []*firestore.DocumentSnapshot
	for {
		doc, err := docs.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	return results, nil
}

type field struct {
	Name  string
	Value any
}

func update[T any](f *Firebase, ctx context.Context, collection, id string, fields ...field) (*T, error) {
	original := new(T)
	err := f.get(ctx, collection, id, original)
	if err != nil {
		return nil, err
	}

	var updates []firestore.Update
	ogV := reflect.ValueOf(original).Elem()
	ogT := reflect.TypeOf(original).Elem()
	for _, newField := range fields {
		ogField := ogV.FieldByName(newField.Name)
		if ogField.Interface() != newField.Value {
			fieldType, ok := ogT.FieldByName(newField.Name)
			if !ok {
				continue
			}
			updates = append(updates, firestore.Update{Path: fieldType.Tag.Get("json"), Value: newField.Value})
			if ogField.CanSet() {
				ogField.Set(reflect.ValueOf(newField.Value))
			}
		}
	}
	if len(updates) == 0 {
		return original, nil
	}

	updatedField := ogV.FieldByName("Updated")
	now := time.Now()
	if updatedField.CanSet() {
		updatedFieldT, ok := ogT.FieldByName("Updated")
		if ok {
			updates = append(updates, firestore.Update{Path: updatedFieldT.Tag.Get("json"), Value: now})
			updatedField.Set(reflect.ValueOf(now))
		}
	}

	_, err = f.fire.Collection(collection).Doc(id).Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return original, nil
}

func (f *Firebase) updateHiddenForAll(ctx context.Context, hidden bool, docs []*firestore.DocumentSnapshot) error {
	for _, doc := range docs {
		_, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "hidden", Value: true}, {Path: "updated", Value: time.Now()}})
		if err != nil {
			return err
		}
	}
	return nil
}

// ////////////
// HELPERS
// ////////////

func docsToType[T any](docs []*firestore.DocumentSnapshot) ([]*T, error) {
	var results []*T
	for _, doc := range docs {
		data := new(T)
		err := doc.DataTo(data)
		if err != nil {
			return nil, err
		}
		results = append(results, data)
	}
	return results, nil
}

// ////////////
// PRICEBOOK
// ////////////

func (f *Firebase) CreatePricebook(ctx context.Context, name, description string) (*Pricebook, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(PricebookCollection).NewDoc()

	data := Pricebook{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		Name:        name,
		Description: description,
		Created:     now,
		Updated:     now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *Firebase) GetPricebook(ctx context.Context, id ID) (*Pricebook, error) {
	data := &Pricebook{}
	err := f.get(ctx, PricebookCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) GetPricebooks(ctx context.Context) ([]*Pricebook, error) {
	docs, err := f.query(ctx, PricebookCollection)
	if err != nil {
		return nil, err
	}
	return docsToType[Pricebook](docs)
}

func (f *Firebase) UpdatePricebook(ctx context.Context, pb Pricebook) (*Pricebook, error) {
	return update[Pricebook](f, ctx, PricebookCollection, pb.Id,
		field{"Name", pb.Name},
		field{"Description", pb.Description},
	)
}

func (f *Firebase) DeletePricebook(ctx context.Context, id ID) error {
	_, err := update[Pricebook](f, ctx, PricebookCollection, id,
		field{"Hidden", true},
	)
	return err
}

func (f *Firebase) RecoverPricebook(ctx context.Context, id ID) error {
	_, err := update[Pricebook](f, ctx, PricebookCollection, id,
		field{"Hidden", false},
	)
	return err
}

// ////////////
// CATEGORY
// ////////////

func (f *Firebase) CreateCategory(ctx context.Context, pricebookId ID, name, description string) (*Category, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(CategoryCollection).NewDoc()

	data := Category{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		PricebookId: pricebookId,
		Name:        name,
		Description: description,
		Created:     now,
		Updated:     now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *Firebase) GetCategory(ctx context.Context, id ID) (*Category, error) {
	data := &Category{}
	err := f.get(ctx, CategoryCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) GetCategories(ctx context.Context, pricebookId ID) ([]*Category, error) {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return nil, err
	}
	return docsToType[Category](docs)
}

func (f *Firebase) UpdateCategoryInfo(ctx context.Context, id ID, name, description string) (*Category, error) {
	return update[Category](f, ctx, CategoryCollection, id,
		field{"Name", name},
		field{"Description", description},
	)
}

func (f *Firebase) UpdateCategoryImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Category, error) {
	return update[Category](f, ctx, CategoryCollection, id,
		field{"ImageId", imageId},
		field{"ThumbnailId", thumbnailId},
	)
}

func (f *Firebase) MoveCategory(ctx context.Context, id ID, parentId ID) (*Category, error) {
	// check that the parentId is a valid collection and is owned by this user
	other := &Category{}
	err := f.get(ctx, CategoryCollection, parentId, other)
	if err != nil {
		return nil, err
	}
	return update[Category](f, ctx, CategoryCollection, id,
		field{"ParentId", parentId},
	)
}

func (f *Firebase) DeleteCategory(ctx context.Context, id ID) error {
	_, err := update[Category](f, ctx, CategoryCollection, id,
		field{"Hidden", true},
	)
	return err
}

func (f *Firebase) DeletePricebookCategories(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) RecoverCategory(ctx context.Context, id ID) error {
	_, err := update[Category](f, ctx, CategoryCollection, id,
		field{"Hidden", false},
	)
	return err
}

func (f *Firebase) RecoverPricebookCategories(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

// ////////////
// ITEMS
// ////////////

func (f *Firebase) CreateItem(ctx context.Context, categoryId ID, name, description string) (*Item, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(ItemCollection).NewDoc()

	data := Item{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		CategoryId:  categoryId,
		Name:        name,
		Description: description,
		Created:     now,
		Updated:     now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *Firebase) GetItem(ctx context.Context, id ID) (*Item, error) {
	data := &Item{}
	err := f.get(ctx, ItemCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) GetSimpleItem(ctx context.Context, id ID) (*SimpleItem, error) {
	// MW: this is not really a good implementation since the point of the SimpleItem is
	// to reduce the amount of processing and data transfer.
	data := &Item{}
	err := f.get(ctx, ItemCollection, id, data)
	if err != nil {
		return nil, err
	}
	return &SimpleItem{
		Id:          data.Id,
		OrgId:       data.OrgId,
		GroupId:     data.GroupId,
		Name:        data.Name,
		ThumbnailId: data.ThumbnailId,
	}, nil
}

func (f *Firebase) GetItemsInCategory(ctx context.Context, categoryId ID) ([]*Item, error) {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "categoryId", Operator: "==", Value: categoryId},
	)
	if err != nil {
		return nil, err
	}
	return docsToType[Item](docs)
}

func (f *Firebase) MoveItem(ctx context.Context, id ID, newCategoryId ID) (*Item, error) {
	// check that the parentId is a valid collection and is owned by this user
	other := &Category{}
	err := f.get(ctx, CategoryCollection, newCategoryId, other)
	if err != nil {
		return nil, err
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"CategoryId", newCategoryId},
	)
}

func (f *Firebase) UpdateItemInfo(ctx context.Context, id ID, code, sku, name, description string) (*Item, error) {
	return update[Item](f, ctx, ItemCollection, id,
		field{"Code", code},
		field{"SKU", sku},
		field{"Name", name},
		field{"Description", description},
	)
}

func (f *Firebase) UpdateItemCost(ctx context.Context, id ID, cost int) (*Item, error) {
	return update[Item](f, ctx, ItemCollection, id,
		field{"Cost", cost},
	)
}

func (f *Firebase) AddItemPrice(ctx context.Context, id ID, priceId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.PriceIds
	for _, elem := range list {
		if elem == priceId {
			// no need to update the database, this list already contains the value
			return orig, nil
		}
	}
	list = append(list, priceId)
	return update[Item](f, ctx, ItemCollection, id,
		field{"PriceIds", list},
	)
}

func (f *Firebase) RemoveItemPrice(ctx context.Context, id ID, priceId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.PriceIds
	found := false
	for i, elem := range list {
		if elem == priceId {
			list = append(list[:i], list[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be removed
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"PriceIds", list},
	)
}

func (f *Firebase) AddItemTag(ctx context.Context, id ID, tagId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.TagIds
	for _, elem := range list {
		if elem == tagId {
			// no need to update the database, this list already contains the value
			return orig, nil
		}
	}
	list = append(list, tagId)
	return update[Item](f, ctx, ItemCollection, id,
		field{"TagIds", list},
	)
}

func (f *Firebase) RemoveItemTag(ctx context.Context, id ID, tagId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.TagIds
	found := false
	for i, elem := range list {
		if elem == tagId {
			list = append(list[:i], list[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be removed
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"TagIds", list},
	)
}

func (f *Firebase) RemoveTagFromItems(ctx context.Context, pricebookId, tagId ID) error {
	docs, err := f.query(ctx, ItemCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
		firestore.PropertyFilter{Path: "tagIds", Operator: "array-contains", Value: tagId},
	)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		_, err = f.RemoveItemTag(ctx, doc.Ref.ID, tagId)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Firebase) UpdateItemHideFromCustomer(ctx context.Context, id ID, hideFromCustomer bool) (*Item, error) {
	return update[Item](f, ctx, ItemCollection, id,
		field{"HideFromCustomer", hideFromCustomer},
	)
}

func (f *Firebase) UpdateItemImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Item, error) {
	return update[Item](f, ctx, ItemCollection, id,
		field{"ImageId", imageId},
		field{"ThumbnailId", thumbnailId},
	)
}

func (f *Firebase) SearchItemsInPricebook(ctx context.Context, pricebookId ID, search string) ([]*Item, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) DeleteItem(ctx context.Context, id ID) error {
	_, err := update[Item](f, ctx, ItemCollection, id,
		field{"Hidden", true},
	)
	return err
}

func (f *Firebase) DeleteCategoryItems(ctx context.Context, categoryId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "categoryId", Operator: "==", Value: categoryId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) DeletePricebookItems(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) RecoverItem(ctx context.Context, id ID) error {
	_, err := update[Item](f, ctx, ItemCollection, id,
		field{"Hidden", false},
	)
	return err
}

func (f *Firebase) RecoverCategoryItems(ctx context.Context, categoryId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "categoryId", Operator: "==", Value: categoryId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

func (f *Firebase) RecoverPricebookItems(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

// ////////////
// SUB ITEM
// ////////////

func (f *Firebase) AddSubItem(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	subItem := &Item{}
	err = f.get(ctx, ItemCollection, subItemId, subItem)
	if err != nil {
		return nil, err
	}
	list := orig.SubItemIds
	for _, elem := range list {
		if elem.SubItemID == subItemId {
			// no need to update the database, this list already contains the value
			return orig, nil
		}
	}
	var firstPriceId *ID
	if len(subItem.PriceIds) > 0 {
		tmp := subItem.PriceIds[0]
		firstPriceId = &tmp
	}
	list = append(list, SubItem{SubItemID: subItemId, Quantity: quantity, PriceId: firstPriceId})
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItemIds", list},
	)
}

func (f *Firebase) UpdateSubItemQuantity(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.SubItemIds
	found := false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			orig.SubItemIds[i].Quantity = quantity
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be updated
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItemIds", list},
	)
}

func (f *Firebase) UpdateSubItemPrice(ctx context.Context, id ID, subItemId ID, priceId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	subItem := &Item{}
	err = f.get(ctx, ItemCollection, subItemId, subItem)
	if err != nil {
		return nil, err
	}
	found := false
	for _, p := range subItem.PriceIds {
		if p == subItemId {
			found = true
			break
		}
	}
	if !found {
		return nil, InvalidPriceIdError
	}

	list := orig.SubItemIds
	found = false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			orig.SubItemIds[i].PriceId = &priceId
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be updated
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItemIds", list},
	)
}

func (f *Firebase) RemoveSubItem(ctx context.Context, id ID, subItemId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := orig.SubItemIds
	found := false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			list = append(list[:i], list[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be removed
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItemIds", list},
	)
}

// ////////////
// PRICE
// ////////////

func (f *Firebase) CreatePrice(ctx context.Context, itemId ID, amount int) (*Price, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	item := &Item{}
	err = f.get(ctx, item.Code, itemId, item)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(PriceCollection).NewDoc()

	data := Price{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		ItemId:      itemId,
		CategoryId:  item.CategoryId,
		PricebookId: item.PricebookId,
		Amount:      amount,
		Created:     now,
		Updated:     now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *Firebase) GetPrice(ctx context.Context, id ID) (*Price, error) {
	data := &Price{}
	err := f.get(ctx, PriceCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) GetPricesByItem(ctx context.Context, itemId ID) ([]*Price, error) {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "itemId", Operator: "==", Value: itemId},
	)
	if err != nil {
		return nil, err
	}
	return docsToType[Price](docs)
}

func (f *Firebase) MovePricesByItem(ctx context.Context, itemId, categoryId ID) error {
	other := &Category{}
	err := f.get(ctx, CategoryCollection, categoryId, other)
	if err != nil {
		return err
	}
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "itemId", Operator: "==", Value: itemId},
	)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		_, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "categoryId", Value: categoryId}, {Path: "updated", Value: time.Now()}})
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Firebase) UpdatePrice(ctx context.Context, p Price) (*Price, error) {
	return update[Price](f, ctx, PriceCollection, p.Id,
		field{Name: "Name", Value: p.Name},
		field{Name: "Description", Value: p.Description},
		field{Name: "Amount", Value: p.Amount},
		field{Name: "Prefix", Value: p.Prefix},
		field{Name: "Suffix", Value: p.Suffix},
	)
}

func (f *Firebase) DeletePrice(ctx context.Context, id ID) error {
	_, err := update[Price](f, ctx, PriceCollection, id,
		field{Name: "Hidden", Value: true},
	)
	return err
}

func (f *Firebase) DeletePricesByItem(ctx context.Context, itemId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "itemId", Operator: "==", Value: itemId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) DeleteCategoryPrices(ctx context.Context, categoryId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "categoryId", Operator: "==", Value: categoryId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) DeletePricebookPrices(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) RecoverPricesByItem(ctx context.Context, itemId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "itemId", Operator: "==", Value: itemId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

func (f *Firebase) RecoverCategoryPrices(ctx context.Context, categoryId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "categoryId", Operator: "==", Value: categoryId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

func (f *Firebase) RecoverPricebookPrices(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, PriceCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

// ////////////
// TAG
// ////////////

func (f *Firebase) CreateTag(ctx context.Context, pricebookId ID, name, description string) (*Tag, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	pb := &Pricebook{}
	err = f.get(ctx, PricebookCollection, pricebookId, pb)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(TagCollection).NewDoc()

	data := Tag{
		Id:              doc.ID,
		OrgId:           orgId,
		GroupId:         groupId,
		PricebookId:     pricebookId,
		Name:            name,
		Description:     description,
		BackgroundColor: "#ffffff",
		TextColor:       "#000000",
		Created:         now,
		Updated:         now,
	}

	_, err = doc.Create(ctx, data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (f *Firebase) GetTag(ctx context.Context, id ID) (*Tag, error) {
	data := &Tag{}
	err := f.get(ctx, TagCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) GetTags(ctx context.Context, pricebookId ID) ([]*Tag, error) {
	docs, err := f.query(ctx, TagCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return nil, err
	}
	return docsToType[Tag](docs)
}

func (f *Firebase) UpdateTagInfo(ctx context.Context, id ID, name, description string) (*Tag, error) {
	return update[Tag](f, ctx, TagCollection, id,
		field{Name: "Name", Value: name},
		field{Name: "Description", Value: description},
	)
}

func (f *Firebase) SearchTags(ctx context.Context, pricebookId ID, search string) ([]*Tag, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) DeleteTag(ctx context.Context, id ID) error {
	_, err := update[Tag](f, ctx, TagCollection, id,
		field{Name: "Hidden", Value: true},
	)
	return err
}

func (f *Firebase) DeletePricebookTags(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, TagCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, true, docs)
}

func (f *Firebase) RecoverPricebookTags(ctx context.Context, pricebookId ID) error {
	docs, err := f.query(ctx, TagCollection,
		firestore.PropertyFilter{Path: "pricebookId", Operator: "==", Value: pricebookId},
	)
	if err != nil {
		return err
	}
	return f.updateHiddenForAll(ctx, false, docs)
}

// ////////////
// IMAGE
// ////////////

func (f *Firebase) CreateImage(ctx context.Context, data []byte) (ID, error) {
	// TODO: implement me
	return "", nil
}

func (f *Firebase) GetImageUrl(ctx context.Context, id ID) (string, error) {
	// TODO: implement me
	return "", nil
}

func (f *Firebase) GetImageBase64(ctx context.Context, id ID) (string, error) {
	// TODO: implement me
	return "", nil
}

func (f *Firebase) GetImageData(ctx context.Context, id ID) ([]byte, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) DeleteImage(ctx context.Context, id ID) error {
	// TODO: implement me
	return nil
}

// ////////////
// QUOTE
// ////////////

func (f *Firebase) CreateQuote(ctx context.Context) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) CreateDuplicateQuote(ctx context.Context, quoteId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) GetQuote(ctx context.Context, id ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteCode(ctx context.Context, id ID, code string) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteOrderNumber(ctx context.Context, id ID, orderNumber string) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteLogoId(ctx context.Context, id ID, logoId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteIssueDate(ctx context.Context, id ID, issueDate *time.Time) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteExpirationDate(ctx context.Context, id ID, expirationDate *time.Time) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuotePaymentTerms(ctx context.Context, id ID, paymentTerms string) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteNotes(ctx context.Context, id ID, notes string) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSenderId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteBillToId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteShipToId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSubTotal(ctx context.Context, id ID, subTotal int) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteTotal(ctx context.Context, id ID, total int) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteBalanceDue(ctx context.Context, id ID, balanceDue int) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteBalancePercentDue(ctx context.Context, id ID, balancePercentDue int) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteBalanceDueOn(ctx context.Context, id ID, balanceDueOn *time.Time) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuotePayUrl(ctx context.Context, id ID, payUrl string) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSent(ctx context.Context, id ID, sent bool) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSentOn(ctx context.Context, id ID, sentOn *time.Time) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSold(ctx context.Context, id ID, sold bool) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateQuoteSoldOn(ctx context.Context, id ID, soldOn *time.Time) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) LockQuote(ctx context.Context, id ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) DeleteQuote(ctx context.Context, id ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) QuoteAddLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) QuoteRemoveLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) QuoteAddAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) QuoteRemoveAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error) {
	// TODO: implement me
	return nil, nil
}

// ////////////
// LINE ITEM
// ////////////

func (f *Firebase) CreateLineItem(ctx context.Context, quoteId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) CreateSubLineItem(ctx context.Context, quoteId, parentId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) CreateDuplicateLineItem(ctx context.Context, id ID) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) GetLineItem(ctx context.Context, id ID) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) MoveLineItem(ctx context.Context, id ID, parentId *ID, index *int) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemImage(ctx context.Context, id ID, imageId *ID) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemDescription(ctx context.Context, id ID, description string) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemQuantity(ctx context.Context, id ID, quantity int, prefix, suffix string) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemUnitPrice(ctx context.Context, id ID, unitPrice int, prefix, suffix string) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemAmount(ctx context.Context, id ID, amount *int, prefix, suffix string) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateLineItemOpen(ctx context.Context, id ID, open bool) (*LineItem, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) DeleteLineItem(ctx context.Context, id ID) error {
	// TODO: implement me
	return nil
}

// ////////////
// ADJUSTMENT
// ////////////

func (f *Firebase) CreateAdjustment(ctx context.Context, quoteId ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) GetAdjustment(ctx context.Context, id ID) (*Adjustment, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) UpdateAdjustment(ctx context.Context, id ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error) {
	// TODO: implement me
	return nil, nil
}

func (f *Firebase) RemoveAdjustment(ctx context.Context, id ID) error {
	// TODO: implement me
	return nil
}

// ////////////
// CONTACT
// ////////////

func (f *Firebase) GetContact(ctx context.Context, id ID) (*Contact, error) {
	// TODO: implement me
	return nil, nil
}

// ////////////
// HELPER
// ////////////

func (f *Firebase) Transaction(ctx context.Context, v func(ctx context.Context) error) error {
	// TODO: implement me
	return nil
}
