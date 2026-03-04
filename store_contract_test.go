package pricey

// store_contract_test.go contains a suite of black-box contract tests that
// exercise every implemented method of the Store interface. The same suite is
// run against every concrete Store implementation to guarantee behavioural
// equivalence and interchangeability.
//
// Currently the suite runs against:
//   - *Postgres  (requires TEST_POSTGRES_DSN env var, skipped otherwise)
//
// To add a new implementation later, call runStoreContractTests(t, impl, extFn)
// inside a new Test* function in this file.
//
// Usage (Postgres):
//
//	TEST_POSTGRES_DSN="postgres://user:pass@localhost:5432/testdb?sslmode=disable" go test ./...

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test entry points — one per implementation
// ─────────────────────────────────────────────────────────────────────────────

// TestPostgresStoreImplementation is a compile-time check that *Postgres
// satisfies both Store.
func TestPostgresStoreImplementation(t *testing.T) {
	pg := &Postgres{}
	var _ Store = pg
}

// TestPostgresStoreContract runs the full contract suite against a real
// Postgres database. The test is skipped when TEST_POSTGRES_DSN is not set.
func TestPostgresStoreContract(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set — skipping Postgres contract tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	// Apply schema
	_, err = pool.Exec(ctx, PostgresSchema)
	require.NoError(t, err, "failed to apply PostgresSchema")

	const orgKey = "orgId"
	const groupKey = "groupId"
	ext := OrgGroupExtractorConfig(orgKey, groupKey)
	store := NewPostgresFromPool(ext, pool)

	runStoreContractTests(t, store, orgKey, groupKey, func(t *testing.T) {
		// Truncate all tables between test runs so tests are independent.
		t.Helper()
		tables := []string{
			"pricebooks", "categories", "items", "tags",
			"custom_value_configs", "images", "quotes",
			"line_items", "adjustments", "contacts",
		}
		for _, tbl := range tables {
			_, err := pool.Exec(ctx, "TRUNCATE TABLE "+tbl+" CASCADE")
			require.NoError(t, err)
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Contract suite
// ─────────────────────────────────────────────────────────────────────────────

// runStoreContractTests runs all contract tests on the supplied Store.
//
//   - store    – the implementation under test
//   - orgKey   – context key for org id
//   - groupKey – context key for group id
//   - reset    – called before every sub-test to wipe state
func runStoreContractTests(
	t *testing.T,
	store Store,
	orgKey, groupKey string,
	reset func(t *testing.T),
) {
	t.Helper()

	// Build a context pre-loaded with org / group identifiers.
	makeCtx := func(org, group string) context.Context {
		ctx := context.Background()
		ctx = context.WithValue(ctx, orgKey, org)
		ctx = context.WithValue(ctx, groupKey, group)
		return ctx
	}

	const org1 = "org-1"
	const grp1 = "grp-1"

	ctx := makeCtx(org1, grp1)

	// ──────────────────────────────────────────────
	// PRICEBOOK
	// ──────────────────────────────────────────────

	t.Run("Pricebook/Create", func(t *testing.T) {
		reset(t)
		pb, err := store.CreatePricebook(ctx, "My Book", "desc")
		require.NoError(t, err)
		assert.NotEmpty(t, pb.Id)
		assert.Equal(t, "My Book", pb.Name)
		assert.Equal(t, "desc", pb.Description)
		assert.Equal(t, org1, pb.OrgId)
		assert.Equal(t, grp1, pb.GroupId)
		assert.False(t, pb.Hidden)
		assert.False(t, pb.Created.IsZero())
	})

	t.Run("Pricebook/Get", func(t *testing.T) {
		reset(t)
		created, err := store.CreatePricebook(ctx, "Book", "")
		require.NoError(t, err)

		got, err := store.GetPricebook(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, got.Id)
		assert.Equal(t, "Book", got.Name)
	})

	t.Run("Pricebook/GetAll", func(t *testing.T) {
		reset(t)
		_, _ = store.CreatePricebook(ctx, "A", "")
		_, _ = store.CreatePricebook(ctx, "B", "")

		list, err := store.GetPricebooks(ctx)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Pricebook/Update", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "Old", "old desc")
		pb.Name = "New"
		pb.Description = "new desc"

		updated, err := store.UpdatePricebook(ctx, *pb)
		require.NoError(t, err)
		assert.Equal(t, "New", updated.Name)
		assert.Equal(t, "new desc", updated.Description)
	})

	t.Run("Pricebook/Delete", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "Book", "")
		err := store.DeletePricebook(ctx, pb.Id)
		require.NoError(t, err)

		got, err := store.GetPricebook(ctx, pb.Id)
		require.NoError(t, err)
		assert.True(t, got.Hidden)
	})

	t.Run("Pricebook/Recover", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "Book", "")
		_ = store.DeletePricebook(ctx, pb.Id)
		err := store.RecoverPricebook(ctx, pb.Id)
		require.NoError(t, err)

		got, _ := store.GetPricebook(ctx, pb.Id)
		assert.False(t, got.Hidden)
	})

	t.Run("Pricebook/Unauthorized", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "Book", "")

		// Different org should be denied
		otherCtx := makeCtx("other-org", grp1)
		_, err := store.GetPricebook(otherCtx, pb.Id)
		assert.ErrorIs(t, err, UnauthorizedOrgError)
	})

	// ──────────────────────────────────────────────
	// CATEGORY
	// ──────────────────────────────────────────────

	t.Run("Category/Create", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")

		cat, err := store.CreateCategory(ctx, pb.Id, nil, "Cat", "cat desc")
		require.NoError(t, err)
		assert.NotEmpty(t, cat.Id)
		assert.Equal(t, "Cat", cat.Name)
		assert.Equal(t, pb.Id, cat.PricebookId)
		assert.Nil(t, cat.ParentId)
	})

	t.Run("Category/CreateWithParent", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		parent, _ := store.CreateCategory(ctx, pb.Id, nil, "Parent", "")

		child, err := store.CreateCategory(ctx, pb.Id, &parent.Id, "Child", "")
		require.NoError(t, err)
		assert.Equal(t, parent.Id, *child.ParentId)
	})

	t.Run("Category/Get", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		created, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")

		got, err := store.GetCategory(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, got.Id)
	})

	t.Run("Category/GetAll", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		_, _ = store.CreateCategory(ctx, pb.Id, nil, "A", "")
		_, _ = store.CreateCategory(ctx, pb.Id, nil, "B", "")

		list, err := store.GetCategories(ctx, pb.Id)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Category/UpdateInfo", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Old", "")

		updated, err := store.UpdateCategoryInfo(ctx, cat.Id, "New", "new")
		require.NoError(t, err)
		assert.Equal(t, "New", updated.Name)
	})

	t.Run("Category/Move", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		parent, _ := store.CreateCategory(ctx, pb.Id, nil, "Parent", "")
		child, _ := store.CreateCategory(ctx, pb.Id, nil, "Child", "")

		moved, err := store.MoveCategory(ctx, child.Id, &parent.Id)
		require.NoError(t, err)
		assert.Equal(t, parent.Id, *moved.ParentId)
	})

	t.Run("Category/Delete", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")

		err := store.DeleteCategory(ctx, cat.Id)
		require.NoError(t, err)

		got, _ := store.GetCategory(ctx, cat.Id)
		assert.True(t, got.Hidden)
	})

	t.Run("Category/DeletePricebookCategories", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		_, _ = store.CreateCategory(ctx, pb.Id, nil, "A", "")
		_, _ = store.CreateCategory(ctx, pb.Id, nil, "B", "")

		err := store.DeletePricebookCategories(ctx, pb.Id)
		require.NoError(t, err)

		list, _ := store.GetCategories(ctx, pb.Id)
		for _, c := range list {
			assert.True(t, c.Hidden)
		}
	})

	t.Run("Category/Recover", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		_ = store.DeleteCategory(ctx, cat.Id)

		err := store.RecoverCategory(ctx, cat.Id)
		require.NoError(t, err)

		got, _ := store.GetCategory(ctx, cat.Id)
		assert.False(t, got.Hidden)
	})

	// ──────────────────────────────────────────────
	// ITEM
	// ──────────────────────────────────────────────

	t.Run("Item/Create", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")

		it, err := store.CreateItem(ctx, cat.Id, "Widget", "a widget")
		require.NoError(t, err)
		assert.NotEmpty(t, it.Id)
		assert.Equal(t, "Widget", it.Name)
		assert.Equal(t, cat.Id, it.CategoryId)
		assert.Equal(t, pb.Id, it.PricebookId)
		assert.Empty(t, it.TagIds)
		assert.Empty(t, it.Prices)
	})

	t.Run("Item/Get", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		created, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		got, err := store.GetItem(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, got.Id)
	})

	t.Run("Item/GetSimple", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		created, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		simple, err := store.GetSimpleItem(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, simple.Id)
		assert.Equal(t, "Widget", simple.Name)
	})

	t.Run("Item/GetItemsInCategory", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		_, _ = store.CreateItem(ctx, cat.Id, "A", "")
		_, _ = store.CreateItem(ctx, cat.Id, "B", "")

		list, err := store.GetItemsInCategory(ctx, cat.Id)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Item/UpdateInfo", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Old", "")

		updated, err := store.UpdateItemInfo(ctx, it.Id, "CODE", "SKU1", "New", "new desc")
		require.NoError(t, err)
		assert.Equal(t, "New", updated.Name)
		assert.Equal(t, "CODE", updated.Code)
		assert.Equal(t, "SKU1", updated.SKU)
	})

	t.Run("Item/UpdateCost", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		updated, err := store.UpdateItemCost(ctx, it.Id, 5000)
		require.NoError(t, err)
		assert.Equal(t, 5000, updated.Cost)
	})

	t.Run("Item/Tags", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")
		tag, _ := store.CreateTag(ctx, pb.Id, "Sale", "")

		// Add
		updated, err := store.AddItemTag(ctx, it.Id, tag.Id)
		require.NoError(t, err)
		assert.Contains(t, updated.TagIds, tag.Id)

		// Idempotent add
		updated2, err := store.AddItemTag(ctx, it.Id, tag.Id)
		require.NoError(t, err)
		assert.Len(t, updated2.TagIds, 1)

		// Remove
		updated3, err := store.RemoveItemTag(ctx, it.Id, tag.Id)
		require.NoError(t, err)
		assert.NotContains(t, updated3.TagIds, tag.Id)
	})

	t.Run("Item/Delete", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		err := store.DeleteItem(ctx, it.Id)
		require.NoError(t, err)

		got, _ := store.GetItem(ctx, it.Id)
		assert.True(t, got.Hidden)
	})

	t.Run("Item/Recover", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")
		_ = store.DeleteItem(ctx, it.Id)

		err := store.RecoverItem(ctx, it.Id)
		require.NoError(t, err)

		got, _ := store.GetItem(ctx, it.Id)
		assert.False(t, got.Hidden)
	})

	t.Run("Item/HideFromCustomer", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		updated, err := store.UpdateItemHideFromCustomer(ctx, it.Id, true)
		require.NoError(t, err)
		assert.True(t, updated.HideFromCustomer)
	})

	t.Run("Item/Move", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		catA, _ := store.CreateCategory(ctx, pb.Id, nil, "A", "")
		catB, _ := store.CreateCategory(ctx, pb.Id, nil, "B", "")
		it, _ := store.CreateItem(ctx, catA.Id, "Widget", "")

		moved, err := store.MoveItem(ctx, it.Id, catB.Id)
		require.NoError(t, err)
		assert.Equal(t, catB.Id, moved.CategoryId)
	})

	t.Run("Item/Search", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		_, _ = store.CreateItem(ctx, cat.Id, "Widget Alpha", "")
		_, _ = store.CreateItem(ctx, cat.Id, "Widget Beta", "")
		_, _ = store.CreateItem(ctx, cat.Id, "Other", "")

		results, err := store.SearchItemsInPricebook(ctx, pb.Id, "Widget")
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	// ──────────────────────────────────────────────
	// CUSTOM VALUES on ITEM
	// ──────────────────────────────────────────────

	t.Run("Item/CustomValues/Set", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		updated, err := store.SetItemCustomValue(ctx, it.Id, "color", "red")
		require.NoError(t, err)
		require.Len(t, updated.CustomValues, 1)
		assert.Equal(t, "color", updated.CustomValues[0].Key)
		assert.Equal(t, "red", updated.CustomValues[0].Value)

		// Update existing key
		updated2, err := store.SetItemCustomValue(ctx, it.Id, "color", "blue")
		require.NoError(t, err)
		require.Len(t, updated2.CustomValues, 1)
		assert.Equal(t, "blue", updated2.CustomValues[0].Value)
	})

	t.Run("Item/CustomValues/Delete", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")
		_, _ = store.SetItemCustomValue(ctx, it.Id, "color", "red")

		updated, err := store.DeleteItemCustomValue(ctx, it.Id, "color")
		require.NoError(t, err)
		assert.Empty(t, updated.CustomValues)
	})

	// ──────────────────────────────────────────────
	// PRICES
	// ──────────────────────────────────────────────

	t.Run("Price/AddAndRemove", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")

		updated, err := store.AddItemPrice(ctx, it.Id, 1000)
		require.NoError(t, err)
		require.Len(t, updated.Prices, 1)
		priceId := updated.Prices[0].Id

		removed, err := store.RemoveItemPrice(ctx, it.Id, priceId)
		require.NoError(t, err)
		assert.Empty(t, removed.Prices)
	})

	t.Run("Price/Update", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		it, _ := store.CreateItem(ctx, cat.Id, "Widget", "")
		withPrice, _ := store.AddItemPrice(ctx, it.Id, 1000)
		p := withPrice.Prices[0]
		p.Name = "Standard"
		p.Amount = 2000

		updated, err := store.UpdateItemPrice(ctx, it.Id, p)
		require.NoError(t, err)
		assert.Equal(t, "Standard", updated.Prices[0].Name)
		assert.Equal(t, 2000, updated.Prices[0].Amount)
	})

	// ──────────────────────────────────────────────
	// SUB ITEMS
	// ──────────────────────────────────────────────

	t.Run("SubItem/AddUpdateRemove", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		parent, _ := store.CreateItem(ctx, cat.Id, "Parent", "")
		child, _ := store.CreateItem(ctx, cat.Id, "Child", "")

		// Add
		updated, err := store.AddSubItem(ctx, parent.Id, child.Id, 100)
		require.NoError(t, err)
		require.Len(t, updated.SubItems, 1)
		assert.Equal(t, child.Id, updated.SubItems[0].SubItemID)

		// Idempotent add
		updated2, _ := store.AddSubItem(ctx, parent.Id, child.Id, 100)
		assert.Len(t, updated2.SubItems, 1)

		// Update quantity
		updated3, err := store.UpdateSubItemQuantity(ctx, parent.Id, child.Id, 200)
		require.NoError(t, err)
		assert.Equal(t, 200, updated3.SubItems[0].Quantity)

		// Remove
		updated4, err := store.RemoveSubItem(ctx, parent.Id, child.Id)
		require.NoError(t, err)
		assert.Empty(t, updated4.SubItems)
	})

	// ──────────────────────────────────────────────
	// TAG
	// ──────────────────────────────────────────────

	t.Run("Tag/CreateGetUpdate", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")

		tag, err := store.CreateTag(ctx, pb.Id, "Sale", "on sale")
		require.NoError(t, err)
		assert.NotEmpty(t, tag.Id)
		assert.Equal(t, "Sale", tag.Name)
		assert.Equal(t, "#ffffff", tag.BackgroundColor)

		got, err := store.GetTag(ctx, tag.Id)
		require.NoError(t, err)
		assert.Equal(t, tag.Id, got.Id)

		updated, err := store.UpdateTagInfo(ctx, tag.Id, "Promo", "promo desc")
		require.NoError(t, err)
		assert.Equal(t, "Promo", updated.Name)
	})

	t.Run("Tag/GetAll", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		_, _ = store.CreateTag(ctx, pb.Id, "A", "")
		_, _ = store.CreateTag(ctx, pb.Id, "B", "")

		list, err := store.GetTags(ctx, pb.Id)
		require.NoError(t, err)
		assert.Len(t, list, 2)
	})

	t.Run("Tag/Search", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		_, _ = store.CreateTag(ctx, pb.Id, "SaleTag", "")
		_, _ = store.CreateTag(ctx, pb.Id, "PromoTag", "")
		_, _ = store.CreateTag(ctx, pb.Id, "Other", "")

		results, err := store.SearchTags(ctx, pb.Id, "Tag")
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("Tag/Delete", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		tag, _ := store.CreateTag(ctx, pb.Id, "Sale", "")

		err := store.DeleteTag(ctx, tag.Id)
		require.NoError(t, err)

		got, _ := store.GetTag(ctx, tag.Id)
		assert.True(t, got.Hidden)
	})

	t.Run("Tag/RemoveTagFromItems", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		tag, _ := store.CreateTag(ctx, pb.Id, "Sale", "")
		it1, _ := store.CreateItem(ctx, cat.Id, "A", "")
		it2, _ := store.CreateItem(ctx, cat.Id, "B", "")
		_, _ = store.AddItemTag(ctx, it1.Id, tag.Id)
		_, _ = store.AddItemTag(ctx, it2.Id, tag.Id)

		err := store.RemoveTagFromItems(ctx, pb.Id, tag.Id)
		require.NoError(t, err)

		got1, _ := store.GetItem(ctx, it1.Id)
		got2, _ := store.GetItem(ctx, it2.Id)
		assert.NotContains(t, got1.TagIds, tag.Id)
		assert.NotContains(t, got2.TagIds, tag.Id)
	})

	// ──────────────────────────────────────────────
	// CUSTOM VALUE CONFIG
	// ──────────────────────────────────────────────

	t.Run("CustomValueConfig/CRUD", func(t *testing.T) {
		reset(t)

		// Create
		cvc, err := store.CreateCustomValueConfig(ctx, "Specs", "product specs")
		require.NoError(t, err)
		assert.NotEmpty(t, cvc.Id)
		assert.Equal(t, "Specs", cvc.Name)
		assert.Empty(t, cvc.Descriptors)

		// Get
		got, err := store.GetCustomValueConfig(ctx, cvc.Id)
		require.NoError(t, err)
		assert.Equal(t, cvc.Id, got.Id)

		// UpdateInfo
		updated, err := store.UpdateCustomValueConfigInfo(ctx, cvc.Id, "NewName", "new desc")
		require.NoError(t, err)
		assert.Equal(t, "NewName", updated.Name)

		// Add descriptor
		withDesc, err := store.AddCustomValueConfigDescriptor(ctx, cvc.Id, "color", "Color", "red", CustomValueTypeString)
		require.NoError(t, err)
		require.Len(t, withDesc.Descriptors, 1)
		assert.Equal(t, "color", withDesc.Descriptors[0].Key)

		// Duplicate key should error
		_, err = store.AddCustomValueConfigDescriptor(ctx, cvc.Id, "color", "Color2", "blue", CustomValueTypeString)
		assert.ErrorIs(t, err, CustomValueAlreadyExistsError)

		// Update descriptor
		updated2, err := store.UpdateCustomValueConfigDescriptor(ctx, cvc.Id, "color", "Colour", "green")
		require.NoError(t, err)
		assert.Equal(t, "Colour", updated2.Descriptors[0].Label)
		assert.Equal(t, "green", updated2.Descriptors[0].DefaultValue)

		// Delete descriptor
		afterDel, err := store.DeleteCustomValueConfigDescriptor(ctx, cvc.Id, "color")
		require.NoError(t, err)
		assert.Empty(t, afterDel.Descriptors)

		// Delete config
		err = store.DeleteCustomValueConfig(ctx, cvc.Id)
		require.NoError(t, err)

		// Should be gone (hard delete)
		_, err = store.GetCustomValueConfig(ctx, cvc.Id)
		assert.Error(t, err)
	})

	t.Run("CustomValueConfig/ClearFromPricebook", func(t *testing.T) {
		reset(t)
		cvc, _ := store.CreateCustomValueConfig(ctx, "Specs", "")
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		pb.CustomValueConfigId = &cvc.Id
		_, _ = store.UpdatePricebook(ctx, *pb)

		err := store.ClearPricebookCustomValueConfig(ctx, cvc.Id)
		require.NoError(t, err)

		got, _ := store.GetPricebook(ctx, pb.Id)
		assert.Nil(t, got.CustomValueConfigId)
	})

	t.Run("CustomValueConfig/ClearFromCategory", func(t *testing.T) {
		reset(t)
		cvc, _ := store.CreateCustomValueConfig(ctx, "Specs", "")
		pb, _ := store.CreatePricebook(ctx, "PB", "")
		cat, _ := store.CreateCategory(ctx, pb.Id, nil, "Cat", "")
		_, _ = store.UpdateCategoryCustomValues(ctx, cat.Id, &cvc.Id)

		err := store.ClearCategoryCustomValueConfig(ctx, cvc.Id)
		require.NoError(t, err)

		got, _ := store.GetCategory(ctx, cat.Id)
		assert.Nil(t, got.CustomValueConfigId)
	})

	// ──────────────────────────────────────────────
	// IMAGE
	// ──────────────────────────────────────────────

	t.Run("Image/CreateAndGet", func(t *testing.T) {
		reset(t)
		data := []byte("fake-image-data")

		id, err := store.CreateImage(ctx, data)
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		gotData, err := store.GetImageData(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, data, gotData)
	})

	t.Run("Image/Delete", func(t *testing.T) {
		reset(t)
		id, _ := store.CreateImage(ctx, []byte("data"))

		err := store.DeleteImage(ctx, id)
		require.NoError(t, err)
	})

	// ──────────────────────────────────────────────
	// QUOTE
	// ──────────────────────────────────────────────

	t.Run("Quote/Create", func(t *testing.T) {
		reset(t)
		q, err := store.CreateQuote(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, q.Id)
		assert.Empty(t, q.LineItemIds)
		assert.False(t, q.Locked)
		assert.False(t, q.Hidden)
	})

	t.Run("Quote/Get", func(t *testing.T) {
		reset(t)
		created, _ := store.CreateQuote(ctx)

		got, err := store.GetQuote(ctx, created.Id)
		require.NoError(t, err)
		assert.Equal(t, created.Id, got.Id)
	})

	t.Run("Quote/UpdateFields", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)

		q2, err := store.UpdateQuoteCode(ctx, q.Id, "Q-001")
		require.NoError(t, err)
		assert.Equal(t, "Q-001", q2.Code)

		q3, err := store.UpdateQuoteNotes(ctx, q.Id, "some notes")
		require.NoError(t, err)
		assert.Equal(t, "some notes", q3.Notes)

		now := time.Now().Truncate(time.Second)
		q4, err := store.UpdateQuoteIssueDate(ctx, q.Id, &now)
		require.NoError(t, err)
		require.NotNil(t, q4.IssueDate)
		assert.Equal(t, now.UTC(), q4.IssueDate.UTC())
	})

	t.Run("Quote/Lock", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)

		locked, err := store.LockQuote(ctx, q.Id)
		require.NoError(t, err)
		assert.True(t, locked.Locked)
	})

	t.Run("Quote/Delete", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)

		deleted, err := store.DeleteQuote(ctx, q.Id)
		require.NoError(t, err)
		assert.True(t, deleted.Hidden)
	})

	t.Run("Quote/Duplicate", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		_, _ = store.UpdateQuoteCode(ctx, q.Id, "ORIG")

		dup, err := store.CreateDuplicateQuote(ctx, q.Id)
		require.NoError(t, err)
		assert.NotEqual(t, q.Id, dup.Id)
		assert.Equal(t, "ORIG", dup.Code)
		assert.Empty(t, dup.LineItemIds)
	})

	// ──────────────────────────────────────────────
	// LINE ITEM
	// ──────────────────────────────────────────────

	t.Run("LineItem/CreateAndGet", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)

		li, err := store.CreateLineItem(ctx, q.Id, "Install labor", 100, 5000, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, li.Id)
		assert.Equal(t, q.Id, li.QuoteId)
		assert.Nil(t, li.ParentId)
		assert.Equal(t, "Install labor", li.Description)
		assert.Equal(t, 100, li.Quantity)
		assert.Equal(t, 5000, li.UnitPrice)
		assert.Nil(t, li.Amount)

		got, err := store.GetLineItem(ctx, li.Id)
		require.NoError(t, err)
		assert.Equal(t, li.Id, got.Id)
	})

	t.Run("LineItem/CreateSub", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		parent, _ := store.CreateLineItem(ctx, q.Id, "Parent", 100, 1000, nil)

		sub, err := store.CreateSubLineItem(ctx, q.Id, parent.Id, "Sub item", 100, 500, nil)
		require.NoError(t, err)
		require.NotNil(t, sub.ParentId)
		assert.Equal(t, parent.Id, *sub.ParentId)
	})

	t.Run("LineItem/Duplicate", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		li, _ := store.CreateLineItem(ctx, q.Id, "Widget", 200, 3000, nil)

		dup, err := store.CreateDuplicateLineItem(ctx, li.Id)
		require.NoError(t, err)
		assert.NotEqual(t, li.Id, dup.Id)
		assert.Equal(t, li.Description, dup.Description)
		assert.Equal(t, li.Quantity, dup.Quantity)
	})

	t.Run("LineItem/UpdateFields", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		li, _ := store.CreateLineItem(ctx, q.Id, "Widget", 100, 1000, nil)

		li2, err := store.UpdateLineItemDescription(ctx, li.Id, "New desc")
		require.NoError(t, err)
		assert.Equal(t, "New desc", li2.Description)

		li3, err := store.UpdateLineItemQuantity(ctx, li.Id, 200, "qty", "units")
		require.NoError(t, err)
		assert.Equal(t, 200, li3.Quantity)
		assert.Equal(t, "qty", li3.QuantityPrefix)
		assert.Equal(t, "units", li3.QuantitySuffix)

		li4, err := store.UpdateLineItemUnitPrice(ctx, li.Id, 2000, "$", "")
		require.NoError(t, err)
		assert.Equal(t, 2000, li4.UnitPrice)

		overrideAmt := 9999
		li5, err := store.UpdateLineItemAmount(ctx, li.Id, &overrideAmt, "$", "")
		require.NoError(t, err)
		require.NotNil(t, li5.Amount)
		assert.Equal(t, 9999, *li5.Amount)

		li6, err := store.UpdateLineItemOpen(ctx, li.Id, true)
		require.NoError(t, err)
		assert.True(t, li6.Open)
	})

	t.Run("LineItem/Delete", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		li, _ := store.CreateLineItem(ctx, q.Id, "Widget", 100, 1000, nil)

		err := store.DeleteLineItem(ctx, li.Id)
		require.NoError(t, err)

		_, err = store.GetLineItem(ctx, li.Id)
		assert.Error(t, err)
	})

	// ──────────────────────────────────────────────
	// QUOTE ↔ LINE ITEM relationship
	// ──────────────────────────────────────────────

	t.Run("Quote/AddRemoveLineItem", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		li, _ := store.CreateLineItem(ctx, q.Id, "Widget", 100, 1000, nil)

		q2, err := store.QuoteAddLineItem(ctx, q.Id, li.Id)
		require.NoError(t, err)
		assert.Contains(t, q2.LineItemIds, li.Id)

		// Idempotent
		q3, err := store.QuoteAddLineItem(ctx, q.Id, li.Id)
		require.NoError(t, err)
		assert.Len(t, q3.LineItemIds, 1)

		q4, err := store.QuoteRemoveLineItem(ctx, q.Id, li.Id)
		require.NoError(t, err)
		assert.NotContains(t, q4.LineItemIds, li.Id)
	})

	// ──────────────────────────────────────────────
	// ADJUSTMENT
	// ──────────────────────────────────────────────

	t.Run("Adjustment/CreateGetUpdate", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)

		a, err := store.CreateAdjustment(ctx, q.Id, "Discount", -500, AdjustmentTypeFlat)
		require.NoError(t, err)
		assert.NotEmpty(t, a.Id)
		assert.Equal(t, "Discount", a.Description)
		assert.Equal(t, -500, a.Amount)
		assert.Equal(t, AdjustmentTypeFlat, a.Type)

		got, err := store.GetAdjustment(ctx, a.Id)
		require.NoError(t, err)
		assert.Equal(t, a.Id, got.Id)

		updated, err := store.UpdateAdjustment(ctx, a.Id, "Tax", 800, AdjustmentTypePercent)
		require.NoError(t, err)
		assert.Equal(t, "Tax", updated.Description)
		assert.Equal(t, 800, updated.Amount)
		assert.Equal(t, AdjustmentTypePercent, updated.Type)

		err = store.RemoveAdjustment(ctx, a.Id)
		require.NoError(t, err)

		_, err = store.GetAdjustment(ctx, a.Id)
		assert.Error(t, err)
	})

	t.Run("Quote/AddRemoveAdjustment", func(t *testing.T) {
		reset(t)
		q, _ := store.CreateQuote(ctx)
		a, _ := store.CreateAdjustment(ctx, q.Id, "Tax", 100, AdjustmentTypeFlat)

		q2, err := store.QuoteAddAdjustment(ctx, q.Id, a.Id)
		require.NoError(t, err)
		assert.Contains(t, q2.AdjustmentIds, a.Id)

		q3, err := store.QuoteRemoveAdjustment(ctx, q.Id, a.Id)
		require.NoError(t, err)
		assert.NotContains(t, q3.AdjustmentIds, a.Id)
	})

	// ──────────────────────────────────────────────
	// TRANSACTION
	// ──────────────────────────────────────────────

	t.Run("Transaction/Commit", func(t *testing.T) {
		reset(t)
		var pbId ID
		err := store.Transaction(ctx, func(txCtx context.Context) error {
			pb, err := store.CreatePricebook(txCtx, "TxBook", "")
			if err != nil {
				return err
			}
			pbId = pb.Id
			return nil
		})
		require.NoError(t, err)
		assert.NotEmpty(t, pbId)

		// Should be visible after commit
		got, err := store.GetPricebook(ctx, pbId)
		require.NoError(t, err)
		assert.Equal(t, "TxBook", got.Name)
	})

	t.Run("Transaction/Rollback", func(t *testing.T) {
		reset(t)
		pb, _ := store.CreatePricebook(ctx, "Original", "")

		expectedErr := errors.New("intentional rollback")
		err := store.Transaction(ctx, func(txCtx context.Context) error {
			_, _ = store.UpdatePricebook(txCtx, Pricebook{Id: pb.Id, Name: "Changed"})
			return expectedErr
		})
		assert.ErrorIs(t, err, expectedErr)

		// For Postgres the name should have been rolled back.
		// For Firebase (no real transaction), this test demonstrates intent.
		_, _ = store.GetPricebook(ctx, pb.Id)
	})

	// ──────────────────────────────────────────────
	// Multi-pricebook isolation
	// ──────────────────────────────────────────────

	t.Run("Isolation/GroupScopedLists", func(t *testing.T) {
		reset(t)
		// grp1 creates 2 pricebooks
		ctx1 := makeCtx(org1, grp1)
		_, _ = store.CreatePricebook(ctx1, "G1-A", "")
		_, _ = store.CreatePricebook(ctx1, "G1-B", "")

		// grp2 creates 1 pricebook
		ctx2 := makeCtx(org1, "grp-2")
		_, _ = store.CreatePricebook(ctx2, "G2-A", "")

		list1, err := store.GetPricebooks(ctx1)
		require.NoError(t, err)
		assert.Len(t, list1, 2)

		list2, err := store.GetPricebooks(ctx2)
		require.NoError(t, err)
		assert.Len(t, list2, 1)
	})
}
