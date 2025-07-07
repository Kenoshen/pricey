package pricey

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"slices"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Firebase struct {
	app  *firebase.App
	auth *auth.Client
	fire *firestore.Client
	ext  OrgGroupExtractor
}

type Collection string

const (
	PricebookCollection   Collection = "pricebook"
	CategoryCollection    Collection = "category"
	ItemCollection        Collection = "item"
	TagCollection         Collection = "tag"
	CustomValueCollection Collection = "customValue"
)

var (
	UnauthorizedOrgError          = errors.New("orgId did not match orgId of document")
	UnauthorizedGroupError        = errors.New("groupId did not match groupId of document")
	InvalidPriceIdError           = errors.New("price id is invalid because it does not exist on the sub item")
	CustomValueAlreadyExistsError = errors.New("custom value keys must be unique")
	CustomValueNotFoundError      = errors.New("no custom value matches the provided key")
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
func (f *Firebase) get(ctx context.Context, collection Collection, id string, dataOut any) error {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return err
	}
	result, err := f.fire.Collection(string(collection)).Doc(id).Get(ctx)
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
func (f *Firebase) query(ctx context.Context, collection Collection, filters ...firestore.EntityFilter) ([]*firestore.DocumentSnapshot, error) {
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
	docs := f.fire.Collection(string(collection)).Query.WhereEntity(firestore.AndFilter{Filters: finalFilters}).Documents(ctx)
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

func update[T any](f *Firebase, ctx context.Context, collection Collection, id string, fields ...field) (*T, error) {
	original := new(T)
	err := f.get(ctx, collection, id, original)
	if err != nil {
		return nil, err
	}

	updates := innerUpdate(original, time.Now(), fields...)
	if len(updates) == 0 {
		return original, nil
	}

	_, err = f.fire.Collection(string(collection)).Doc(id).Update(ctx, updates)
	if err != nil {
		return nil, err
	}
	return original, nil
}

func innerUpdate[T any](original *T, now time.Time, fields ...field) []firestore.Update {
	var updates []firestore.Update
	ogV := reflect.ValueOf(original).Elem()
	ogT := reflect.TypeOf(original).Elem()
	for _, newField := range fields {
		ogField := ogV.FieldByName(newField.Name)
		a := ogField.Interface()
		b := newField.Value
		equal := false
		switch ogField.Kind() {
		case reflect.Slice:
			equal = reflect.DeepEqual(a, b)
		default:
			equal = a == b
		}
		if !equal {
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
		return nil
	}

	updatedField := ogV.FieldByName("Updated")
	if updatedField.CanSet() {
		updatedFieldT, ok := ogT.FieldByName("Updated")
		if ok {
			updates = append(updates, firestore.Update{Path: updatedFieldT.Tag.Get("json"), Value: now})
			updatedField.Set(reflect.ValueOf(now))
		}
	}
	return updates
}

func (f *Firebase) updateHiddenForAll(ctx context.Context, hidden bool, docs []*firestore.DocumentSnapshot) error {
	for _, doc := range docs {
		_, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "hidden", Value: hidden}, {Path: "updated", Value: time.Now()}})
		if err != nil {
			return err
		}
	}
	return nil
}

// delete actually deletes the document out of the database,
// not just setting hidden to true
func (f *Firebase) delete(ctx context.Context, collection Collection, id ID) error {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return err
	}
	result, err := f.fire.Collection(string(collection)).Doc(id).Get(ctx)
	if err != nil {
		return err
	}
	dataOut := map[string]any{}
	err = result.DataTo(&dataOut)
	if err != nil {
		return err
	}
	if dataOut["orgId"] != orgId {
		return UnauthorizedOrgError
	}
	if dataOut["groupId"] != groupId {
		return UnauthorizedGroupError
	}
	_, err = result.Ref.Delete(ctx)
	return err
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
	doc := f.fire.Collection(string(PricebookCollection)).NewDoc()

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
		field{"CustomValueConfigId", pb.CustomValueConfigId},
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

func (f *Firebase) ClearPricebookCustomValueConfig(ctx context.Context, configId ID) error {
	docs, err := f.query(ctx, PricebookCollection,
		firestore.PropertyFilter{Path: "customValueConfigId", Operator: "==", Value: configId},
	)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		_, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "customValueConfigId", Value: nil}, {Path: "updated", Value: time.Now()}})
		if err != nil {
			return err
		}
	}
	return nil
}

// ////////////
// CATEGORY
// ////////////

func (f *Firebase) CreateCategory(ctx context.Context, pricebookId ID, parentId *ID, name, description string) (*Category, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	pb := Pricebook{}
	err = f.get(ctx, PricebookCollection, pricebookId, &pb)
	if err != nil {
		return nil, err
	}
	if parentId != nil {
		parent := Category{}
		err = f.get(ctx, CategoryCollection, *parentId, &parent)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	doc := f.fire.Collection(string(CategoryCollection)).NewDoc()

	data := Category{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		PricebookId: pricebookId,
		ParentId:    parentId,
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

func (f *Firebase) UpdateCategoryCustomValues(ctx context.Context, id ID, customValuesId *ID) (*Category, error) {
	return update[Category](f, ctx, CategoryCollection, id,
		field{"CustomValueConfigId", customValuesId},
	)
}

func (f *Firebase) MoveCategory(ctx context.Context, id ID, parentId *ID) (*Category, error) {
	// check that the parentId is a valid collection and is owned by this user
	if parentId != nil {
		other := &Category{}
		err := f.get(ctx, CategoryCollection, *parentId, other)
		if err != nil {
			return nil, err
		}
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

func (f *Firebase) ClearCategoryCustomValueConfig(ctx context.Context, configId ID) error {
	docs, err := f.query(ctx, CategoryCollection,
		firestore.PropertyFilter{Path: "customValueConfigId", Operator: "==", Value: configId},
	)
	if err != nil {
		return err
	}
	for _, doc := range docs {
		_, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "customValueConfigId", Value: nil}, {Path: "updated", Value: time.Now()}})
		if err != nil {
			return err
		}
	}
	return nil
}

// ////////////
// ITEMS
// ////////////

func (f *Firebase) CreateItem(ctx context.Context, categoryId ID, name, description string) (*Item, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	category := Category{}
	err = f.get(ctx, CategoryCollection, categoryId, &category)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(string(ItemCollection)).NewDoc()

	data := Item{
		Id:          doc.ID,
		OrgId:       orgId,
		GroupId:     groupId,
		CategoryId:  categoryId,
		PricebookId: category.PricebookId,
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

func (f *Firebase) AddItemTag(ctx context.Context, id ID, tagId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.TagIds)
	if slices.Contains(list, tagId) {
		// no need to update the database, this list already contains the value
		return orig, nil
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
	list := slices.Clone(orig.TagIds)
	found := false
	for i, elem := range list {
		if elem == tagId {
			list = slices.Delete(list, i, i+1)
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

func (f *Firebase) SetItemCustomValue(ctx context.Context, itemId ID, key ID, value string) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.CustomValues)
	found := false
	for i, elem := range list {
		if elem.Key == key {
			list[i].Value = value
			found = true
		}
	}
	if !found {
		list = append(list, CustomValue{
			Key:   key,
			Value: value,
		})
	}
	return update[Item](f, ctx, ItemCollection, itemId,
		field{"CustomValues", list},
	)
}

func (f *Firebase) DeleteItemCustomValue(ctx context.Context, itemId ID, key ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.CustomValues)
	found := false
	for i, elem := range list {
		if elem.Key == key {
			list = slices.Delete(list, i, i+1)
			found = true
			break
		}
	}
	if !found {
		// if the custom value doesn't exist, then just return
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, itemId,
		field{"CustomValues", list},
	)
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
	for _, elem := range orig.SubItems {
		if elem.SubItemID == subItemId {
			// no need to update the database, this list already contains the value
			return orig, nil
		}
	}
	subItem := &Item{}
	err = f.get(ctx, ItemCollection, subItemId, subItem)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.SubItems)
	var firstPriceId *ID
	if len(subItem.Prices) > 0 {
		tmp := subItem.Prices[0]
		firstPriceId = &(tmp.Id)
	}
	list = append(list, SubItem{SubItemID: subItemId, Quantity: quantity, PriceId: firstPriceId})
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItems", list},
	)
}

func (f *Firebase) UpdateSubItemQuantity(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.SubItems)
	found := false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			list[i].Quantity = quantity
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be updated
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItems", list},
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
	for _, p := range subItem.Prices {
		if p.Id == subItemId {
			found = true
			break
		}
	}
	if !found {
		return nil, InvalidPriceIdError
	}

	list := slices.Clone(orig.SubItems)
	found = false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			list[i].PriceId = &priceId
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be updated
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItems", list},
	)
}

func (f *Firebase) RemoveSubItem(ctx context.Context, id ID, subItemId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.SubItems)
	found := false
	for i, elem := range list {
		if elem.SubItemID == subItemId {
			list = slices.Delete(list, i, i+1)
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be removed
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"SubItems", list},
	)
}

// ////////////
// PRICE
// ////////////

func (f *Firebase) AddItemPrice(ctx context.Context, itemId ID, amount int) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	list := slices.Clone(orig.Prices)
	list = append(list, Price{
		Id:          uuid.NewString(),
		Name:        "",
		Description: "",
		Amount:      amount,
		Prefix:      "",
		Suffix:      "",
		Created:     now,
		Updated:     now,
	})
	return update[Item](f, ctx, ItemCollection, itemId,
		field{"Prices", list},
	)
}

func (f *Firebase) SetDefaultItemPrice(ctx context.Context, itemId ID, priceId ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Prices)
	index := -1
	var price Price
	for i, elem := range list {
		if elem.Id == priceId {
			elem = price
			list = slices.Delete(list, i, i+1)
			break
		}
	}
	if index <= 0 {
		// the value was not in the list or it was already the default, do nothing
		return orig, nil
	}
	// being the 0th element in the array means you are the default price
	list = slices.Insert(list, 0, price)
	return update[Item](f, ctx, ItemCollection, itemId,
		field{"Prices", list},
	)
}

func (f *Firebase) UpdateItemPrice(ctx context.Context, itemId ID, p Price) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Prices)
	found := false
	for i, elem := range list {
		if elem.Id == p.Id {
			list[i] = Price{
				Id:          elem.Id,
				Name:        p.Name,
				Description: p.Description,
				Amount:      p.Amount,
				Prefix:      p.Prefix,
				Suffix:      p.Suffix,
				Created:     elem.Created,
				Updated:     time.Now(),
			}
			found = true
			break
		}
	}
	if !found {
		// the id did not match any price in the list, do nothing
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, itemId,
		field{"Prices", list},
	)
}

func (f *Firebase) RemoveItemPrice(ctx context.Context, itemId ID, id ID) (*Item, error) {
	orig := &Item{}
	err := f.get(ctx, ItemCollection, itemId, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Prices)
	found := false
	for i, elem := range list {
		if elem.Id == id {
			list = slices.Delete(list, i, i+1)
			found = true
			break
		}
	}
	if !found {
		// if the value was not in the list, it does not need to be removed
		return orig, nil
	}
	return update[Item](f, ctx, ItemCollection, id,
		field{"Prices", list},
	)
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
	doc := f.fire.Collection(string(TagCollection)).NewDoc()

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
// CUSTOM VALUES
// ////////////

func (f *Firebase) CreateCustomValueConfig(ctx context.Context, name string, description string) (*CustomValueConfig, error) {
	orgId, groupId, err := f.ext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := f.fire.Collection(string(CustomValueCollection)).NewDoc()

	data := CustomValueConfig{
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

func (f *Firebase) GetCustomValueConfig(ctx context.Context, id ID) (*CustomValueConfig, error) {
	data := &CustomValueConfig{}
	err := f.get(ctx, CustomValueCollection, id, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (f *Firebase) UpdateCustomValueConfigInfo(ctx context.Context, id ID, name string, description string) (*CustomValueConfig, error) {
	return update[CustomValueConfig](f, ctx, CustomValueCollection, id,
		field{Name: "Name", Value: name},
		field{Name: "Description", Value: description},
	)
}

func (f *Firebase) AddCustomValueConfigDescriptor(ctx context.Context, id ID, key ID, label string, defaultValue string, valueType CustomValueType) (*CustomValueConfig, error) {
	orig := &CustomValueConfig{}
	err := f.get(ctx, CustomValueCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Descriptors)
	for _, elem := range list {
		if elem.Key == key {
			return nil, CustomValueAlreadyExistsError
		}
	}
	list = append(list, CustomValueDescriptor{
		Key:          key,
		Label:        label,
		DefaultValue: defaultValue,
		ValueType:    valueType,
	})
	return update[CustomValueConfig](f, ctx, CustomValueCollection, id,
		field{"Descriptors", list},
	)
}

func (f *Firebase) UpdateCustomValueConfigDescriptor(ctx context.Context, id ID, key ID, label string, defaultValue string) (*CustomValueConfig, error) {
	orig := &CustomValueConfig{}
	err := f.get(ctx, CustomValueCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Descriptors)
	found := false
	for i, elem := range list {
		if elem.Key == key {
			list[i].Label = label
			list[i].DefaultValue = defaultValue
			found = true
		}
	}
	if !found {
		return nil, CustomValueNotFoundError
	}
	return update[CustomValueConfig](f, ctx, CustomValueCollection, id,
		field{"Descriptors", list},
	)
}

func (f *Firebase) DeleteCustomValueConfigDescriptor(ctx context.Context, id ID, key ID) (*CustomValueConfig, error) {
	orig := &CustomValueConfig{}
	err := f.get(ctx, CustomValueCollection, id, orig)
	if err != nil {
		return nil, err
	}
	list := slices.Clone(orig.Descriptors)
	found := false
	for i, elem := range list {
		if elem.Key == key {
			list = slices.Delete(list, i, i+1)
			found = true
		}
	}
	if !found {
		return orig, nil
	}
	return update[CustomValueConfig](f, ctx, CustomValueCollection, id,
		field{"Descriptors", list},
	)
}

func (f *Firebase) DeleteCustomValueConfig(ctx context.Context, id ID) error {
	return f.delete(ctx, CustomValueCollection, id)
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
	// f.fire.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
	// })
	return v(ctx)
}
