package pricey

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres implements the Store interface using a PostgreSQL database.
// Arrays and nested structs (SubItems, Prices, TagIds, CustomValues, Descriptors,
// LineItemIds, AdjustmentIds) are stored as JSONB columns. All columns that need
// to be filtered on (orgId, groupId, pricebookId, etc.) are also stored as
// dedicated text columns so they can be indexed.
type Postgres struct {
	db  *pgxpool.Pool
	ext OrgGroupExtractor
}

// NewPostgres creates a new Postgres store. The caller is responsible for
// running the schema migration (see PostgresSchema) before using the store.
func NewPostgres(ctx context.Context, ext OrgGroupExtractor, connStr string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return &Postgres{db: pool, ext: ext}, nil
}

// NewPostgresFromPool creates a Postgres store from an existing pgxpool.Pool.
func NewPostgresFromPool(ext OrgGroupExtractor, pool *pgxpool.Pool) *Postgres {
	return &Postgres{db: pool, ext: ext}
}

// Close releases the underlying connection pool.
func (p *Postgres) Close() {
	p.db.Close()
}

// PostgresSchema returns the SQL DDL required to set up all tables used by the
// Postgres store. It is safe to run multiple times (all statements use
// CREATE TABLE IF NOT EXISTS).
const PostgresSchema = `
CREATE TABLE IF NOT EXISTS pricebooks (
    id                    TEXT PRIMARY KEY,
    org_id                TEXT NOT NULL,
    group_id              TEXT NOT NULL,
    custom_value_config_id TEXT,
    name                  TEXT NOT NULL DEFAULT '',
    description           TEXT NOT NULL DEFAULT '',
    image_id              TEXT NOT NULL DEFAULT '',
    thumbnail_id          TEXT NOT NULL DEFAULT '',
    created               TIMESTAMPTZ NOT NULL,
    updated               TIMESTAMPTZ NOT NULL,
    hidden                BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS categories (
    id                    TEXT PRIMARY KEY,
    org_id                TEXT NOT NULL,
    group_id              TEXT NOT NULL,
    parent_id             TEXT,
    pricebook_id          TEXT NOT NULL,
    custom_value_config_id TEXT,
    name                  TEXT NOT NULL DEFAULT '',
    description           TEXT NOT NULL DEFAULT '',
    hide_from_customer    BOOLEAN NOT NULL DEFAULT FALSE,
    image_id              TEXT NOT NULL DEFAULT '',
    thumbnail_id          TEXT NOT NULL DEFAULT '',
    created               TIMESTAMPTZ NOT NULL,
    updated               TIMESTAMPTZ NOT NULL,
    hidden                BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS items (
    id                  TEXT PRIMARY KEY,
    org_id              TEXT NOT NULL,
    group_id            TEXT NOT NULL,
    pricebook_id        TEXT NOT NULL,
    category_id         TEXT NOT NULL,
    parent_ids          JSONB NOT NULL DEFAULT '[]',
    tag_ids             JSONB NOT NULL DEFAULT '[]',
    sub_items           JSONB NOT NULL DEFAULT '[]',
    prices              JSONB NOT NULL DEFAULT '[]',
    code                TEXT NOT NULL DEFAULT '',
    sku                 TEXT NOT NULL DEFAULT '',
    name                TEXT NOT NULL DEFAULT '',
    description         TEXT NOT NULL DEFAULT '',
    cost                INTEGER NOT NULL DEFAULT 0,
    hide_from_customer  BOOLEAN NOT NULL DEFAULT FALSE,
    image_id            TEXT NOT NULL DEFAULT '',
    thumbnail_id        TEXT NOT NULL DEFAULT '',
    created             TIMESTAMPTZ NOT NULL,
    updated             TIMESTAMPTZ NOT NULL,
    hidden              BOOLEAN NOT NULL DEFAULT FALSE,
    custom_values       JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS tags (
    id               TEXT PRIMARY KEY,
    org_id           TEXT NOT NULL,
    group_id         TEXT NOT NULL,
    pricebook_id     TEXT NOT NULL,
    name             TEXT NOT NULL DEFAULT '',
    description      TEXT NOT NULL DEFAULT '',
    background_color TEXT NOT NULL DEFAULT '#ffffff',
    text_color       TEXT NOT NULL DEFAULT '#000000',
    created          TIMESTAMPTZ NOT NULL,
    updated          TIMESTAMPTZ NOT NULL,
    hidden           BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS custom_value_configs (
    id          TEXT PRIMARY KEY,
    org_id      TEXT NOT NULL,
    group_id    TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    descriptors JSONB NOT NULL DEFAULT '[]',
    created     TIMESTAMPTZ NOT NULL,
    updated     TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS images (
    id       TEXT PRIMARY KEY,
    org_id   TEXT NOT NULL,
    group_id TEXT NOT NULL,
    data     BYTEA,
    base64   TEXT NOT NULL DEFAULT '',
    url      TEXT NOT NULL DEFAULT '',
    created  TIMESTAMPTZ NOT NULL,
    hidden   BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS quotes (
    id                       TEXT PRIMARY KEY,
    org_id                   TEXT NOT NULL,
    group_id                 TEXT NOT NULL,
    code                     TEXT NOT NULL DEFAULT '',
    order_number             TEXT NOT NULL DEFAULT '',
    logo_id                  TEXT NOT NULL DEFAULT '',
    primary_background_color TEXT NOT NULL DEFAULT '',
    primary_text_color       TEXT NOT NULL DEFAULT '',
    issue_date               TIMESTAMPTZ,
    expiration_date          TIMESTAMPTZ,
    payment_terms            TEXT NOT NULL DEFAULT '',
    notes                    TEXT NOT NULL DEFAULT '',
    sender_id                TEXT NOT NULL DEFAULT '',
    bill_to_id               TEXT NOT NULL DEFAULT '',
    ship_to_id               TEXT NOT NULL DEFAULT '',
    line_item_ids            JSONB NOT NULL DEFAULT '[]',
    sub_total                INTEGER NOT NULL DEFAULT 0,
    adjustment_ids           JSONB NOT NULL DEFAULT '[]',
    total                    INTEGER NOT NULL DEFAULT 0,
    balance_due              INTEGER NOT NULL DEFAULT 0,
    balance_percent_due      INTEGER NOT NULL DEFAULT 0,
    balance_due_on           TIMESTAMPTZ,
    pay_url                  TEXT NOT NULL DEFAULT '',
    sent                     BOOLEAN NOT NULL DEFAULT FALSE,
    sent_on                  TIMESTAMPTZ,
    sold                     BOOLEAN NOT NULL DEFAULT FALSE,
    sold_on                  TIMESTAMPTZ,
    created                  TIMESTAMPTZ NOT NULL,
    updated                  TIMESTAMPTZ NOT NULL,
    hidden                   BOOLEAN NOT NULL DEFAULT FALSE,
    locked                   BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS line_items (
    id               TEXT PRIMARY KEY,
    org_id           TEXT NOT NULL,
    group_id         TEXT NOT NULL,
    quote_id         TEXT NOT NULL,
    parent_id        TEXT,
    sub_item_ids     JSONB NOT NULL DEFAULT '[]',
    image_id         TEXT,
    description      TEXT NOT NULL DEFAULT '',
    quantity         INTEGER NOT NULL DEFAULT 0,
    quantity_suffix  TEXT NOT NULL DEFAULT '',
    quantity_prefix  TEXT NOT NULL DEFAULT '',
    unit_price       INTEGER NOT NULL DEFAULT 0,
    unit_price_suffix TEXT NOT NULL DEFAULT '',
    unit_price_prefix TEXT NOT NULL DEFAULT '',
    amount           INTEGER,
    amount_suffix    TEXT NOT NULL DEFAULT '',
    amount_prefix    TEXT NOT NULL DEFAULT '',
    open             BOOLEAN NOT NULL DEFAULT FALSE,
    created          TIMESTAMPTZ NOT NULL,
    updated          TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS adjustments (
    id          TEXT PRIMARY KEY,
    org_id      TEXT NOT NULL,
    group_id    TEXT NOT NULL,
    quote_id    TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    type        INTEGER NOT NULL DEFAULT 0,
    amount      INTEGER NOT NULL DEFAULT 0,
    created     TIMESTAMPTZ NOT NULL,
    updated     TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS contacts (
    id           TEXT PRIMARY KEY,
    org_id       TEXT NOT NULL,
    group_id     TEXT NOT NULL,
    name         TEXT NOT NULL DEFAULT '',
    company_name TEXT NOT NULL DEFAULT '',
    phones       JSONB NOT NULL DEFAULT '[]',
    emails       JSONB NOT NULL DEFAULT '[]',
    websites     JSONB NOT NULL DEFAULT '[]',
    street       TEXT NOT NULL DEFAULT '',
    city         TEXT NOT NULL DEFAULT '',
    state        TEXT NOT NULL DEFAULT '',
    zip          TEXT NOT NULL DEFAULT ''
);
`

// ─────────────────────────────────────────────
// internal helpers
// ─────────────────────────────────────────────

func newID() ID {
	return uuid.NewString()
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func mustUnmarshal[T any](b []byte, dst *T) error {
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, dst)
}

// pgxRowToPricebook scans a pgx row into a Pricebook.
func pgxRowToPricebook(row pgx.CollectableRow) (*Pricebook, error) {
	var pb Pricebook
	var customValueConfigId *string
	err := row.Scan(
		&pb.Id, &pb.OrgId, &pb.GroupId, &customValueConfigId,
		&pb.Name, &pb.Description, &pb.ImageId, &pb.ThumbnailId,
		&pb.Created, &pb.Updated, &pb.Hidden,
	)
	if err != nil {
		return nil, err
	}
	pb.CustomValueConfigId = customValueConfigId
	return &pb, nil
}

func pgxRowToCategory(row pgx.CollectableRow) (*Category, error) {
	var c Category
	var parentId, customValueConfigId *string
	err := row.Scan(
		&c.Id, &c.OrgId, &c.GroupId, &parentId, &c.PricebookId, &customValueConfigId,
		&c.Name, &c.Description, &c.HideFromCustomer,
		&c.ImageId, &c.ThumbnailId, &c.Created, &c.Updated, &c.Hidden,
	)
	if err != nil {
		return nil, err
	}
	c.ParentId = parentId
	c.CustomValueConfigId = customValueConfigId
	return &c, nil
}

func pgxRowToItem(row pgx.CollectableRow) (*Item, error) {
	var it Item
	var parentIdsJSON, tagIdsJSON, subItemsJSON, pricesJSON, customValuesJSON []byte
	err := row.Scan(
		&it.Id, &it.OrgId, &it.GroupId, &it.PricebookId, &it.CategoryId,
		&parentIdsJSON, &tagIdsJSON, &subItemsJSON, &pricesJSON,
		&it.Code, &it.SKU, &it.Name, &it.Description, &it.Cost,
		&it.HideFromCustomer, &it.ImageId, &it.ThumbnailId,
		&it.Created, &it.Updated, &it.Hidden, &customValuesJSON,
	)
	if err != nil {
		return nil, err
	}
	_ = mustUnmarshal(parentIdsJSON, &it.ParentIds)
	_ = mustUnmarshal(tagIdsJSON, &it.TagIds)
	_ = mustUnmarshal(subItemsJSON, &it.SubItems)
	_ = mustUnmarshal(pricesJSON, &it.Prices)
	_ = mustUnmarshal(customValuesJSON, &it.CustomValues)
	if it.ParentIds == nil {
		it.ParentIds = []ID{}
	}
	if it.TagIds == nil {
		it.TagIds = []ID{}
	}
	if it.SubItems == nil {
		it.SubItems = []SubItem{}
	}
	if it.Prices == nil {
		it.Prices = []Price{}
	}
	if it.CustomValues == nil {
		it.CustomValues = []CustomValue{}
	}
	return &it, nil
}

func pgxRowToTag(row pgx.CollectableRow) (*Tag, error) {
	var t Tag
	err := row.Scan(
		&t.Id, &t.OrgId, &t.GroupId, &t.PricebookId,
		&t.Name, &t.Description, &t.BackgroundColor, &t.TextColor,
		&t.Created, &t.Updated, &t.Hidden,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func pgxRowToCustomValueConfig(row pgx.CollectableRow) (*CustomValueConfig, error) {
	var c CustomValueConfig
	var descJSON []byte
	err := row.Scan(
		&c.Id, &c.OrgId, &c.GroupId, &c.Name, &c.Description,
		&descJSON, &c.Created, &c.Updated,
	)
	if err != nil {
		return nil, err
	}
	_ = mustUnmarshal(descJSON, &c.Descriptors)
	if c.Descriptors == nil {
		c.Descriptors = []CustomValueDescriptor{}
	}
	return &c, nil
}

func pgxRowToQuote(row pgx.CollectableRow) (*Quote, error) {
	var q Quote
	var lineItemIdsJSON, adjustmentIdsJSON []byte
	err := row.Scan(
		&q.Id, &q.Code, &q.OrderNumber, &q.LogoId,
		&q.PrimaryBackgroundColor, &q.PrimaryTextColor,
		&q.IssueDate, &q.ExpirationDate,
		&q.PaymentTerms, &q.Notes,
		&q.SenderId, &q.BillToId, &q.ShipToId,
		&lineItemIdsJSON, &q.SubTotal, &adjustmentIdsJSON, &q.Total,
		&q.BalanceDue, &q.BalancePercentDue, &q.BalanceDueOn,
		&q.PayUrl, &q.Sent, &q.SentOn, &q.Sold, &q.SoldOn,
		&q.Created, &q.Updated, &q.Hidden, &q.Locked,
	)
	if err != nil {
		return nil, err
	}
	_ = mustUnmarshal(lineItemIdsJSON, &q.LineItemIds)
	_ = mustUnmarshal(adjustmentIdsJSON, &q.AdjustmentIds)
	if q.LineItemIds == nil {
		q.LineItemIds = []ID{}
	}
	if q.AdjustmentIds == nil {
		q.AdjustmentIds = []ID{}
	}
	return &q, nil
}

func pgxRowToLineItem(row pgx.CollectableRow) (*LineItem, error) {
	var li LineItem
	var subItemIdsJSON []byte
	err := row.Scan(
		&li.Id, &li.QuoteId, &li.ParentId, &subItemIdsJSON, &li.ImageId,
		&li.Description, &li.Quantity, &li.QuantitySuffix, &li.QuantityPrefix,
		&li.UnitPrice, &li.UnitPriceSuffix, &li.UnitPricePrefix,
		&li.Amount, &li.AmountSuffix, &li.AmountPrefix,
		&li.Open, &li.Created, &li.Updated,
	)
	if err != nil {
		return nil, err
	}
	_ = mustUnmarshal(subItemIdsJSON, &li.SubItemIds)
	if li.SubItemIds == nil {
		li.SubItemIds = []ID{}
	}
	return &li, nil
}

func pgxRowToAdjustment(row pgx.CollectableRow) (*Adjustment, error) {
	var a Adjustment
	err := row.Scan(
		&a.Id, &a.QuoteId, &a.Description, &a.Type, &a.Amount,
		&a.Created, &a.Updated,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func pgxRowToContact(row pgx.CollectableRow) (*Contact, error) {
	var c Contact
	var phonesJSON, emailsJSON, websitesJSON []byte
	err := row.Scan(
		&c.Id, &c.Name, &c.CompanyName,
		&phonesJSON, &emailsJSON, &websitesJSON,
		&c.Street, &c.City, &c.State, &c.Zip,
	)
	if err != nil {
		return nil, err
	}
	_ = mustUnmarshal(phonesJSON, &c.Phones)
	_ = mustUnmarshal(emailsJSON, &c.Emails)
	_ = mustUnmarshal(websitesJSON, &c.Websites)
	return &c, nil
}

// authCheck verifies that org_id and group_id on the retrieved row match the
// context values. Returns UnauthorizedOrgError / UnauthorizedGroupError on
// mismatch.
func (p *Postgres) authCheck(ctx context.Context, rowOrgId, rowGroupId string) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	if rowOrgId != orgId {
		return UnauthorizedOrgError
	}
	if rowGroupId != groupId {
		return UnauthorizedGroupError
	}
	return nil
}

// ─────────────────────────────────────────────
// PRICEBOOK
// ─────────────────────────────────────────────

const pricebookCols = `id, org_id, group_id, custom_value_config_id, name, description, image_id, thumbnail_id, created, updated, hidden`

func (p *Postgres) CreatePricebook(ctx context.Context, name, description string) (*Pricebook, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO pricebooks (id, org_id, group_id, custom_value_config_id, name, description, image_id, thumbnail_id, created, updated, hidden)
		VALUES ($1,$2,$3,NULL,$4,$5,'',' ',  $6,$7,FALSE)`,
		id, orgId, groupId, name, description, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Pricebook{
		Id: id, OrgId: orgId, GroupId: groupId,
		Name: name, Description: description,
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) GetPricebook(ctx context.Context, id ID) (*Pricebook, error) {
	rows, err := p.db.Query(ctx,
		`SELECT `+pricebookCols+` FROM pricebooks WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	pb, err := pgx.CollectOneRow(rows, pgxRowToPricebook)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, pb.OrgId, pb.GroupId); err != nil {
		return nil, err
	}
	return pb, nil
}

func (p *Postgres) GetPricebooks(ctx context.Context) ([]*Pricebook, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+pricebookCols+` FROM pricebooks WHERE org_id=$1 AND group_id=$2`,
		orgId, groupId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToPricebook)
}

func (p *Postgres) UpdatePricebook(ctx context.Context, pb Pricebook) (*Pricebook, error) {
	existing, err := p.GetPricebook(ctx, pb.Id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `
		UPDATE pricebooks SET name=$1, description=$2, custom_value_config_id=$3, updated=$4
		WHERE id=$5`,
		pb.Name, pb.Description, pb.CustomValueConfigId, now, pb.Id,
	)
	if err != nil {
		return nil, err
	}
	existing.Name = pb.Name
	existing.Description = pb.Description
	existing.CustomValueConfigId = pb.CustomValueConfigId
	existing.Updated = now
	return existing, nil
}

func (p *Postgres) DeletePricebook(ctx context.Context, id ID) error {
	_, err := p.GetPricebook(ctx, id)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE pricebooks SET hidden=TRUE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) RecoverPricebook(ctx context.Context, id ID) error {
	_, err := p.GetPricebook(ctx, id)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE pricebooks SET hidden=FALSE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) ClearPricebookCustomValueConfig(ctx context.Context, configId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `
		UPDATE pricebooks SET custom_value_config_id=NULL, updated=$1
		WHERE org_id=$2 AND group_id=$3 AND custom_value_config_id=$4`,
		time.Now(), orgId, groupId, configId,
	)
	return err
}

// ─────────────────────────────────────────────
// CATEGORY
// ─────────────────────────────────────────────

const categoryCols = `id, org_id, group_id, parent_id, pricebook_id, custom_value_config_id, name, description, hide_from_customer, image_id, thumbnail_id, created, updated, hidden`

func (p *Postgres) CreateCategory(ctx context.Context, pricebookId ID, parentId *ID, name, description string) (*Category, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	// validate parent ownership
	if _, err := p.GetPricebook(ctx, pricebookId); err != nil {
		return nil, err
	}
	if parentId != nil {
		if _, err := p.GetCategory(ctx, *parentId); err != nil {
			return nil, err
		}
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO categories (id, org_id, group_id, parent_id, pricebook_id, custom_value_config_id, name, description, hide_from_customer, image_id, thumbnail_id, created, updated, hidden)
		VALUES ($1,$2,$3,$4,$5,NULL,$6,$7,FALSE,'','', $8,$9,FALSE)`,
		id, orgId, groupId, parentId, pricebookId, name, description, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Category{
		Id: id, OrgId: orgId, GroupId: groupId,
		PricebookId: pricebookId, ParentId: parentId,
		Name: name, Description: description,
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) GetCategory(ctx context.Context, id ID) (*Category, error) {
	rows, err := p.db.Query(ctx,
		`SELECT `+categoryCols+` FROM categories WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	c, err := pgx.CollectOneRow(rows, pgxRowToCategory)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, c.OrgId, c.GroupId); err != nil {
		return nil, err
	}
	return c, nil
}

func (p *Postgres) GetCategories(ctx context.Context, pricebookId ID) ([]*Category, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+categoryCols+` FROM categories WHERE org_id=$1 AND group_id=$2 AND pricebook_id=$3`,
		orgId, groupId, pricebookId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToCategory)
}

func (p *Postgres) UpdateCategoryInfo(ctx context.Context, id ID, name, description string) (*Category, error) {
	c, err := p.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE categories SET name=$1, description=$2, updated=$3 WHERE id=$4`,
		name, description, now, id)
	if err != nil {
		return nil, err
	}
	c.Name = name
	c.Description = description
	c.Updated = now
	return c, nil
}

func (p *Postgres) UpdateCategoryImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Category, error) {
	c, err := p.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE categories SET image_id=$1, thumbnail_id=$2, updated=$3 WHERE id=$4`,
		imageId, thumbnailId, now, id)
	if err != nil {
		return nil, err
	}
	c.ImageId = imageId
	c.ThumbnailId = thumbnailId
	c.Updated = now
	return c, nil
}

func (p *Postgres) UpdateCategoryCustomValues(ctx context.Context, id ID, customValuesId *ID) (*Category, error) {
	c, err := p.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE categories SET custom_value_config_id=$1, updated=$2 WHERE id=$3`,
		customValuesId, now, id)
	if err != nil {
		return nil, err
	}
	c.CustomValueConfigId = customValuesId
	c.Updated = now
	return c, nil
}

func (p *Postgres) MoveCategory(ctx context.Context, id ID, parentId *ID) (*Category, error) {
	if parentId != nil {
		if _, err := p.GetCategory(ctx, *parentId); err != nil {
			return nil, err
		}
	}
	c, err := p.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE categories SET parent_id=$1, updated=$2 WHERE id=$3`,
		parentId, now, id)
	if err != nil {
		return nil, err
	}
	c.ParentId = parentId
	c.Updated = now
	return c, nil
}

func (p *Postgres) DeleteCategory(ctx context.Context, id ID) error {
	if _, err := p.GetCategory(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE categories SET hidden=TRUE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) DeletePricebookCategories(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE categories SET hidden=TRUE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

func (p *Postgres) RecoverCategory(ctx context.Context, id ID) error {
	if _, err := p.GetCategory(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE categories SET hidden=FALSE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) RecoverPricebookCategories(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE categories SET hidden=FALSE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

func (p *Postgres) ClearCategoryCustomValueConfig(ctx context.Context, configId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `
		UPDATE categories SET custom_value_config_id=NULL, updated=$1
		WHERE org_id=$2 AND group_id=$3 AND custom_value_config_id=$4`,
		time.Now(), orgId, groupId, configId)
	return err
}

// ─────────────────────────────────────────────
// ITEMS
// ─────────────────────────────────────────────

const itemCols = `id, org_id, group_id, pricebook_id, category_id, parent_ids, tag_ids, sub_items, prices, code, sku, name, description, cost, hide_from_customer, image_id, thumbnail_id, created, updated, hidden, custom_values`

func (p *Postgres) CreateItem(ctx context.Context, categoryId ID, name, description string) (*Item, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	cat, err := p.GetCategory(ctx, categoryId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	empty := mustMarshal([]any{})
	_, err = p.db.Exec(ctx, `
		INSERT INTO items (id, org_id, group_id, pricebook_id, category_id, parent_ids, tag_ids, sub_items, prices, code, sku, name, description, cost, hide_from_customer, image_id, thumbnail_id, created, updated, hidden, custom_values)
		VALUES ($1,$2,$3,$4,$5,$6,$6,$6,$6,'','', $7,$8,0,FALSE,'','',$9,$10,FALSE,$6)`,
		id, orgId, groupId, cat.PricebookId, categoryId, empty, name, description, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Item{
		Id: id, OrgId: orgId, GroupId: groupId,
		PricebookId: cat.PricebookId, CategoryId: categoryId,
		Name: name, Description: description,
		ParentIds: []ID{}, TagIds: []ID{}, SubItems: []SubItem{},
		Prices: []Price{}, CustomValues: []CustomValue{},
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) GetItem(ctx context.Context, id ID) (*Item, error) {
	rows, err := p.db.Query(ctx, `SELECT `+itemCols+` FROM items WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	it, err := pgx.CollectOneRow(rows, pgxRowToItem)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, it.OrgId, it.GroupId); err != nil {
		return nil, err
	}
	return it, nil
}

func (p *Postgres) GetSimpleItem(ctx context.Context, id ID) (*SimpleItem, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SimpleItem{
		Id:          it.Id,
		OrgId:       it.OrgId,
		GroupId:     it.GroupId,
		Name:        it.Name,
		ThumbnailId: it.ThumbnailId,
	}, nil
}

func (p *Postgres) GetItemsInCategory(ctx context.Context, categoryId ID) ([]*Item, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+itemCols+` FROM items WHERE org_id=$1 AND group_id=$2 AND category_id=$3`,
		orgId, groupId, categoryId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToItem)
}

func (p *Postgres) MoveItem(ctx context.Context, id ID, newCategoryId ID) (*Item, error) {
	if _, err := p.GetCategory(ctx, newCategoryId); err != nil {
		return nil, err
	}
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET category_id=$1, updated=$2 WHERE id=$3`,
		newCategoryId, now, id)
	if err != nil {
		return nil, err
	}
	it.CategoryId = newCategoryId
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateItemInfo(ctx context.Context, id ID, code, sku, name, description string) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET code=$1, sku=$2, name=$3, description=$4, updated=$5 WHERE id=$6`,
		code, sku, name, description, now, id)
	if err != nil {
		return nil, err
	}
	it.Code = code
	it.SKU = sku
	it.Name = name
	it.Description = description
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateItemCost(ctx context.Context, id ID, cost int) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET cost=$1, updated=$2 WHERE id=$3`, cost, now, id)
	if err != nil {
		return nil, err
	}
	it.Cost = cost
	it.Updated = now
	return it, nil
}

func (p *Postgres) AddItemTag(ctx context.Context, id ID, tagId ID) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	if slices.Contains(it.TagIds, tagId) {
		return it, nil
	}
	it.TagIds = append(it.TagIds, tagId)
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET tag_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.TagIds), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) RemoveItemTag(ctx context.Context, id ID, tagId ID) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	found := false
	for i, t := range it.TagIds {
		if t == tagId {
			it.TagIds = slices.Delete(it.TagIds, i, i+1)
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET tag_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.TagIds), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) RemoveTagFromItems(ctx context.Context, pricebookId, tagId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+itemCols+` FROM items WHERE org_id=$1 AND group_id=$2 AND pricebook_id=$3 AND tag_ids @> $4`,
		orgId, groupId, pricebookId, mustMarshal([]ID{tagId}))
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, pgxRowToItem)
	if err != nil {
		return err
	}
	for _, it := range items {
		if _, err := p.RemoveItemTag(ctx, it.Id, tagId); err != nil {
			return err
		}
	}
	return nil
}

func (p *Postgres) UpdateItemHideFromCustomer(ctx context.Context, id ID, hideFromCustomer bool) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET hide_from_customer=$1, updated=$2 WHERE id=$3`,
		hideFromCustomer, now, id)
	if err != nil {
		return nil, err
	}
	it.HideFromCustomer = hideFromCustomer
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateItemImage(ctx context.Context, id ID, imageId, thumbnailId ID) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET image_id=$1, thumbnail_id=$2, updated=$3 WHERE id=$4`,
		imageId, thumbnailId, now, id)
	if err != nil {
		return nil, err
	}
	it.ImageId = imageId
	it.ThumbnailId = thumbnailId
	it.Updated = now
	return it, nil
}

func (p *Postgres) SearchItemsInPricebook(ctx context.Context, pricebookId ID, search string) ([]*Item, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+itemCols+` FROM items WHERE org_id=$1 AND group_id=$2 AND pricebook_id=$3 AND (name ILIKE $4 OR description ILIKE $4 OR code ILIKE $4 OR sku ILIKE $4)`,
		orgId, groupId, pricebookId, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToItem)
}

func (p *Postgres) DeleteItem(ctx context.Context, id ID) error {
	if _, err := p.GetItem(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE items SET hidden=TRUE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) DeleteCategoryItems(ctx context.Context, categoryId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE items SET hidden=TRUE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND category_id=$4`,
		time.Now(), orgId, groupId, categoryId)
	return err
}

func (p *Postgres) DeletePricebookItems(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE items SET hidden=TRUE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

func (p *Postgres) RecoverItem(ctx context.Context, id ID) error {
	if _, err := p.GetItem(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE items SET hidden=FALSE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) RecoverCategoryItems(ctx context.Context, categoryId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE items SET hidden=FALSE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND category_id=$4`,
		time.Now(), orgId, groupId, categoryId)
	return err
}

func (p *Postgres) RecoverPricebookItems(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE items SET hidden=FALSE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

func (p *Postgres) SetItemCustomValue(ctx context.Context, itemId ID, key ID, value string) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	found := false
	for i, cv := range it.CustomValues {
		if cv.Key == key {
			it.CustomValues[i].Value = value
			found = true
		}
	}
	if !found {
		it.CustomValues = append(it.CustomValues, CustomValue{Key: key, Value: value})
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET custom_values=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.CustomValues), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) DeleteItemCustomValue(ctx context.Context, itemId ID, key ID) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	found := false
	for i, cv := range it.CustomValues {
		if cv.Key == key {
			it.CustomValues = slices.Delete(it.CustomValues, i, i+1)
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET custom_values=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.CustomValues), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

// ─────────────────────────────────────────────
// SUB ITEM
// ─────────────────────────────────────────────

func (p *Postgres) AddSubItem(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, s := range it.SubItems {
		if s.SubItemID == subItemId {
			return it, nil
		}
	}
	sub, err := p.GetItem(ctx, subItemId)
	if err != nil {
		return nil, err
	}
	var firstPriceId *ID
	if len(sub.Prices) > 0 {
		tmp := sub.Prices[0].Id
		firstPriceId = &tmp
	}
	it.SubItems = append(it.SubItems, SubItem{SubItemID: subItemId, Quantity: quantity, PriceId: firstPriceId})
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET sub_items=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.SubItems), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateSubItemQuantity(ctx context.Context, id ID, subItemId ID, quantity int) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	found := false
	for i, s := range it.SubItems {
		if s.SubItemID == subItemId {
			it.SubItems[i].Quantity = quantity
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET sub_items=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.SubItems), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateSubItemPrice(ctx context.Context, id ID, subItemId ID, priceId ID) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	sub, err := p.GetItem(ctx, subItemId)
	if err != nil {
		return nil, err
	}
	found := false
	for _, pr := range sub.Prices {
		if pr.Id == priceId {
			found = true
			break
		}
	}
	if !found {
		return nil, InvalidPriceIdError
	}
	found = false
	for i, s := range it.SubItems {
		if s.SubItemID == subItemId {
			it.SubItems[i].PriceId = &priceId
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET sub_items=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.SubItems), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) RemoveSubItem(ctx context.Context, id ID, subItemId ID) (*Item, error) {
	it, err := p.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	found := false
	for i, s := range it.SubItems {
		if s.SubItemID == subItemId {
			it.SubItems = slices.Delete(it.SubItems, i, i+1)
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET sub_items=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.SubItems), now, id)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

// ─────────────────────────────────────────────
// PRICE
// ─────────────────────────────────────────────

func (p *Postgres) AddItemPrice(ctx context.Context, itemId ID, amount int) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	it.Prices = append(it.Prices, Price{
		Id: uuid.NewString(), Amount: amount,
		Created: now, Updated: now,
	})
	_, err = p.db.Exec(ctx, `UPDATE items SET prices=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.Prices), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) SetDefaultItemPrice(ctx context.Context, itemId ID, priceId ID) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	index := -1
	var price Price
	for i, pr := range it.Prices {
		if pr.Id == priceId {
			price = pr
			index = i
			break
		}
	}
	if index <= 0 {
		return it, nil
	}
	it.Prices = slices.Delete(it.Prices, index, index+1)
	it.Prices = slices.Insert(it.Prices, 0, price)
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET prices=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.Prices), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) UpdateItemPrice(ctx context.Context, itemId ID, pr Price) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	found := false
	for i, existing := range it.Prices {
		if existing.Id == pr.Id {
			it.Prices[i] = Price{
				Id:          existing.Id,
				Name:        pr.Name,
				Description: pr.Description,
				Amount:      pr.Amount,
				Prefix:      pr.Prefix,
				Suffix:      pr.Suffix,
				Created:     existing.Created,
				Updated:     time.Now(),
			}
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET prices=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.Prices), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

func (p *Postgres) RemoveItemPrice(ctx context.Context, itemId ID, id ID) (*Item, error) {
	it, err := p.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}
	found := false
	for i, pr := range it.Prices {
		if pr.Id == id {
			it.Prices = slices.Delete(it.Prices, i, i+1)
			found = true
			break
		}
	}
	if !found {
		return it, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE items SET prices=$1, updated=$2 WHERE id=$3`,
		mustMarshal(it.Prices), now, itemId)
	if err != nil {
		return nil, err
	}
	it.Updated = now
	return it, nil
}

// ─────────────────────────────────────────────
// TAG
// ─────────────────────────────────────────────

const tagCols = `id, org_id, group_id, pricebook_id, name, description, background_color, text_color, created, updated, hidden`

func (p *Postgres) CreateTag(ctx context.Context, pricebookId ID, name, description string) (*Tag, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := p.GetPricebook(ctx, pricebookId); err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO tags (id, org_id, group_id, pricebook_id, name, description, background_color, text_color, created, updated, hidden)
		VALUES ($1,$2,$3,$4,$5,$6,'#ffffff','#000000',$7,$8,FALSE)`,
		id, orgId, groupId, pricebookId, name, description, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Tag{
		Id: id, OrgId: orgId, GroupId: groupId, PricebookId: pricebookId,
		Name: name, Description: description,
		BackgroundColor: "#ffffff", TextColor: "#000000",
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) GetTag(ctx context.Context, id ID) (*Tag, error) {
	rows, err := p.db.Query(ctx, `SELECT `+tagCols+` FROM tags WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	t, err := pgx.CollectOneRow(rows, pgxRowToTag)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, t.OrgId, t.GroupId); err != nil {
		return nil, err
	}
	return t, nil
}

func (p *Postgres) GetTags(ctx context.Context, pricebookId ID) ([]*Tag, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+tagCols+` FROM tags WHERE org_id=$1 AND group_id=$2 AND pricebook_id=$3`,
		orgId, groupId, pricebookId)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToTag)
}

func (p *Postgres) UpdateTagInfo(ctx context.Context, id ID, name, description string) (*Tag, error) {
	t, err := p.GetTag(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE tags SET name=$1, description=$2, updated=$3 WHERE id=$4`,
		name, description, now, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Description = description
	t.Updated = now
	return t, nil
}

func (p *Postgres) SearchTags(ctx context.Context, pricebookId ID, search string) ([]*Tag, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(ctx,
		`SELECT `+tagCols+` FROM tags WHERE org_id=$1 AND group_id=$2 AND pricebook_id=$3 AND (name ILIKE $4 OR description ILIKE $4)`,
		orgId, groupId, pricebookId, "%"+search+"%")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgxRowToTag)
}

func (p *Postgres) DeleteTag(ctx context.Context, id ID) error {
	if _, err := p.GetTag(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE tags SET hidden=TRUE, updated=$1 WHERE id=$2`, time.Now(), id)
	return err
}

func (p *Postgres) DeletePricebookTags(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE tags SET hidden=TRUE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

func (p *Postgres) RecoverPricebookTags(ctx context.Context, pricebookId ID) error {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return err
	}
	_, err = p.db.Exec(ctx, `UPDATE tags SET hidden=FALSE, updated=$1 WHERE org_id=$2 AND group_id=$3 AND pricebook_id=$4`,
		time.Now(), orgId, groupId, pricebookId)
	return err
}

// ─────────────────────────────────────────────
// CUSTOM VALUE CONFIG
// ─────────────────────────────────────────────

const cvcCols = `id, org_id, group_id, name, description, descriptors, created, updated`

func (p *Postgres) CreateCustomValueConfig(ctx context.Context, name string, description string) (*CustomValueConfig, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO custom_value_configs (id, org_id, group_id, name, description, descriptors, created, updated)
		VALUES ($1,$2,$3,$4,$5,'[]',$6,$7)`,
		id, orgId, groupId, name, description, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &CustomValueConfig{
		Id: id, OrgId: orgId, GroupId: groupId,
		Name: name, Description: description,
		Descriptors: []CustomValueDescriptor{},
		Created:     now, Updated: now,
	}, nil
}

func (p *Postgres) GetCustomValueConfig(ctx context.Context, id ID) (*CustomValueConfig, error) {
	rows, err := p.db.Query(ctx, `SELECT `+cvcCols+` FROM custom_value_configs WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	c, err := pgx.CollectOneRow(rows, pgxRowToCustomValueConfig)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, c.OrgId, c.GroupId); err != nil {
		return nil, err
	}
	return c, nil
}

func (p *Postgres) UpdateCustomValueConfigInfo(ctx context.Context, id ID, name string, description string) (*CustomValueConfig, error) {
	c, err := p.GetCustomValueConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE custom_value_configs SET name=$1, description=$2, updated=$3 WHERE id=$4`,
		name, description, now, id)
	if err != nil {
		return nil, err
	}
	c.Name = name
	c.Description = description
	c.Updated = now
	return c, nil
}

func (p *Postgres) AddCustomValueConfigDescriptor(ctx context.Context, id ID, key ID, label string, defaultValue string, valueType CustomValueType) (*CustomValueConfig, error) {
	c, err := p.GetCustomValueConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, d := range c.Descriptors {
		if d.Key == key {
			return nil, CustomValueAlreadyExistsError
		}
	}
	c.Descriptors = append(c.Descriptors, CustomValueDescriptor{
		Key:          key,
		Label:        label,
		DefaultValue: defaultValue,
		ValueType:    valueType,
	})
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE custom_value_configs SET descriptors=$1, updated=$2 WHERE id=$3`,
		mustMarshal(c.Descriptors), now, id)
	if err != nil {
		return nil, err
	}
	c.Updated = now
	return c, nil
}

func (p *Postgres) UpdateCustomValueConfigDescriptor(ctx context.Context, id ID, key ID, label string, defaultValue string) (*CustomValueConfig, error) {
	c, err := p.GetCustomValueConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	found := false
	for i, d := range c.Descriptors {
		if d.Key == key {
			c.Descriptors[i].Label = label
			c.Descriptors[i].DefaultValue = defaultValue
			found = true
		}
	}
	if !found {
		return nil, CustomValueNotFoundError
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE custom_value_configs SET descriptors=$1, updated=$2 WHERE id=$3`,
		mustMarshal(c.Descriptors), now, id)
	if err != nil {
		return nil, err
	}
	c.Updated = now
	return c, nil
}

func (p *Postgres) DeleteCustomValueConfigDescriptor(ctx context.Context, id ID, key ID) (*CustomValueConfig, error) {
	c, err := p.GetCustomValueConfig(ctx, id)
	if err != nil {
		return nil, err
	}
	found := false
	for i, d := range c.Descriptors {
		if d.Key == key {
			c.Descriptors = slices.Delete(c.Descriptors, i, i+1)
			found = true
			break
		}
	}
	if !found {
		return c, nil
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE custom_value_configs SET descriptors=$1, updated=$2 WHERE id=$3`,
		mustMarshal(c.Descriptors), now, id)
	if err != nil {
		return nil, err
	}
	c.Updated = now
	return c, nil
}

func (p *Postgres) DeleteCustomValueConfig(ctx context.Context, id ID) error {
	if _, err := p.GetCustomValueConfig(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `DELETE FROM custom_value_configs WHERE id=$1`, id)
	return err
}

// ─────────────────────────────────────────────
// IMAGE
// ─────────────────────────────────────────────

func (p *Postgres) CreateImage(ctx context.Context, data []byte) (ID, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return "", err
	}
	id := newID()
	now := time.Now()
	_, err = p.db.Exec(ctx, `
		INSERT INTO images (id, org_id, group_id, data, base64, url, created, hidden)
		VALUES ($1,$2,$3,$4,'',' ',$5,FALSE)`,
		id, orgId, groupId, data, now,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (p *Postgres) GetImageUrl(ctx context.Context, id ID) (string, error) {
	img, err := p.getImage(ctx, id)
	if err != nil {
		return "", err
	}
	return img.Url, nil
}

func (p *Postgres) GetImageBase64(ctx context.Context, id ID) (string, error) {
	img, err := p.getImage(ctx, id)
	if err != nil {
		return "", err
	}
	return img.Base64, nil
}

func (p *Postgres) GetImageData(ctx context.Context, id ID) ([]byte, error) {
	img, err := p.getImage(ctx, id)
	if err != nil {
		return nil, err
	}
	return img.Data, nil
}

func (p *Postgres) DeleteImage(ctx context.Context, id ID) error {
	if _, err := p.getImage(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `UPDATE images SET hidden=TRUE WHERE id=$1`, id)
	return err
}

func (p *Postgres) getImage(ctx context.Context, id ID) (*Image, error) {
	var img Image
	rows, err := p.db.Query(ctx, `SELECT id, org_id, group_id, data, base64, url, created, hidden FROM images WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	img2, err := pgx.CollectOneRow(rows, func(row pgx.CollectableRow) (*Image, error) {
		var i Image
		if err := row.Scan(&i.Id, &i.OrgId, &i.GroupId, &i.Data, &i.Base64, &i.Url, &i.Created, &i.Hidden); err != nil {
			return nil, err
		}
		return &i, nil
	})
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, img2.OrgId, img2.GroupId); err != nil {
		return nil, err
	}
	_ = img
	return img2, nil
}

// ─────────────────────────────────────────────
// QUOTE
// ─────────────────────────────────────────────

const quoteCols = `id, code, order_number, logo_id, primary_background_color, primary_text_color, issue_date, expiration_date, payment_terms, notes, sender_id, bill_to_id, ship_to_id, line_item_ids, sub_total, adjustment_ids, total, balance_due, balance_percent_due, balance_due_on, pay_url, sent, sent_on, sold, sold_on, created, updated, hidden, locked`

func (p *Postgres) CreateQuote(ctx context.Context) (*Quote, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO quotes (id, org_id, group_id, code, order_number, logo_id, primary_background_color, primary_text_color, issue_date, expiration_date, payment_terms, notes, sender_id, bill_to_id, ship_to_id, line_item_ids, sub_total, adjustment_ids, total, balance_due, balance_percent_due, balance_due_on, pay_url, sent, sent_on, sold, sold_on, created, updated, hidden, locked)
		VALUES ($1,$2,$3,'','','','','',NULL,NULL,'','','','','',$4,0,$4,0,0,0,NULL,'',FALSE,NULL,FALSE,NULL,$5,$6,FALSE,FALSE)`,
		id, orgId, groupId, mustMarshal([]ID{}), now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Quote{
		Id: id, LineItemIds: []ID{}, AdjustmentIds: []ID{},
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) CreateDuplicateQuote(ctx context.Context, quoteId ID) (*Quote, error) {
	original, err := p.getQuote(ctx, quoteId)
	if err != nil {
		return nil, err
	}
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO quotes (id, org_id, group_id, code, order_number, logo_id, primary_background_color, primary_text_color, issue_date, expiration_date, payment_terms, notes, sender_id, bill_to_id, ship_to_id, line_item_ids, sub_total, adjustment_ids, total, balance_due, balance_percent_due, balance_due_on, pay_url, sent, sent_on, sold, sold_on, created, updated, hidden, locked)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,FALSE,NULL,FALSE,NULL,$24,$25,FALSE,FALSE)`,
		id, orgId, groupId,
		original.Code, original.OrderNumber, original.LogoId,
		original.PrimaryBackgroundColor, original.PrimaryTextColor,
		original.IssueDate, original.ExpirationDate,
		original.PaymentTerms, original.Notes,
		original.SenderId, original.BillToId, original.ShipToId,
		mustMarshal([]ID{}), // reset line items
		original.SubTotal,
		mustMarshal([]ID{}), // reset adjustments
		original.Total, original.BalanceDue, original.BalancePercentDue,
		original.BalanceDueOn, original.PayUrl,
		now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Quote{
		Id:                     id,
		Code:                   original.Code,
		OrderNumber:            original.OrderNumber,
		LogoId:                 original.LogoId,
		PrimaryBackgroundColor: original.PrimaryBackgroundColor,
		PrimaryTextColor:       original.PrimaryTextColor,
		IssueDate:              original.IssueDate,
		ExpirationDate:         original.ExpirationDate,
		PaymentTerms:           original.PaymentTerms,
		Notes:                  original.Notes,
		SenderId:               original.SenderId,
		BillToId:               original.BillToId,
		ShipToId:               original.ShipToId,
		LineItemIds:            []ID{},
		SubTotal:               original.SubTotal,
		AdjustmentIds:          []ID{},
		Total:                  original.Total,
		BalanceDue:             original.BalanceDue,
		BalancePercentDue:      original.BalancePercentDue,
		BalanceDueOn:           original.BalanceDueOn,
		PayUrl:                 original.PayUrl,
		Created:                now,
		Updated:                now,
	}, nil
}

func (p *Postgres) GetQuote(ctx context.Context, id ID) (*Quote, error) {
	return p.getQuote(ctx, id)
}

func (p *Postgres) getQuote(ctx context.Context, id ID) (*Quote, error) {
	rows, err := p.db.Query(ctx,
		`SELECT `+quoteCols+` FROM quotes WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	q, err := pgx.CollectOneRow(rows, pgxRowToQuote)
	if err != nil {
		return nil, err
	}
	// auth check via org_id/group_id stored on the row (we need to fetch them separately)
	var orgId, groupId string
	err = p.db.QueryRow(ctx, `SELECT org_id, group_id FROM quotes WHERE id=$1`, id).Scan(&orgId, &groupId)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, orgId, groupId); err != nil {
		return nil, err
	}
	return q, nil
}

func (p *Postgres) updateQuoteField(ctx context.Context, id ID, col string, val any) (*Quote, error) {
	if _, err := p.getQuote(ctx, id); err != nil {
		return nil, err
	}
	now := time.Now()
	// We use fmt.Sprintf-style but safely: col is a hard-coded string in all callers.
	_, err := p.db.Exec(ctx, `UPDATE quotes SET `+col+`=$1, updated=$2 WHERE id=$3`, val, now, id)
	if err != nil {
		return nil, err
	}
	return p.getQuote(ctx, id)
}

func (p *Postgres) UpdateQuoteCode(ctx context.Context, id ID, code string) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "code", code)
}
func (p *Postgres) UpdateQuoteOrderNumber(ctx context.Context, id ID, orderNumber string) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "order_number", orderNumber)
}
func (p *Postgres) UpdateQuoteLogoId(ctx context.Context, id ID, logoId ID) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "logo_id", logoId)
}
func (p *Postgres) UpdateQuoteIssueDate(ctx context.Context, id ID, issueDate *time.Time) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "issue_date", issueDate)
}
func (p *Postgres) UpdateQuoteExpirationDate(ctx context.Context, id ID, expirationDate *time.Time) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "expiration_date", expirationDate)
}
func (p *Postgres) UpdateQuotePaymentTerms(ctx context.Context, id ID, paymentTerms string) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "payment_terms", paymentTerms)
}
func (p *Postgres) UpdateQuoteNotes(ctx context.Context, id ID, notes string) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "notes", notes)
}
func (p *Postgres) UpdateQuoteSenderId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sender_id", contactId)
}
func (p *Postgres) UpdateQuoteBillToId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "bill_to_id", contactId)
}
func (p *Postgres) UpdateQuoteShipToId(ctx context.Context, id ID, contactId ID) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "ship_to_id", contactId)
}
func (p *Postgres) UpdateQuoteSubTotal(ctx context.Context, id ID, subTotal int) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sub_total", subTotal)
}
func (p *Postgres) UpdateQuoteTotal(ctx context.Context, id ID, total int) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "total", total)
}
func (p *Postgres) UpdateQuoteBalanceDue(ctx context.Context, id ID, balanceDue int) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "balance_due", balanceDue)
}
func (p *Postgres) UpdateQuoteBalancePercentDue(ctx context.Context, id ID, balancePercentDue int) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "balance_percent_due", balancePercentDue)
}
func (p *Postgres) UpdateQuoteBalanceDueOn(ctx context.Context, id ID, balanceDueOn *time.Time) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "balance_due_on", balanceDueOn)
}
func (p *Postgres) UpdateQuotePayUrl(ctx context.Context, id ID, payUrl string) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "pay_url", payUrl)
}
func (p *Postgres) UpdateQuoteSent(ctx context.Context, id ID, sent bool) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sent", sent)
}
func (p *Postgres) UpdateQuoteSentOn(ctx context.Context, id ID, sentOn *time.Time) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sent_on", sentOn)
}
func (p *Postgres) UpdateQuoteSold(ctx context.Context, id ID, sold bool) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sold", sold)
}
func (p *Postgres) UpdateQuoteSoldOn(ctx context.Context, id ID, soldOn *time.Time) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "sold_on", soldOn)
}

func (p *Postgres) LockQuote(ctx context.Context, id ID) (*Quote, error) {
	return p.updateQuoteField(ctx, id, "locked", true)
}

func (p *Postgres) DeleteQuote(ctx context.Context, id ID) (*Quote, error) {
	q, err := p.getQuote(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE quotes SET hidden=TRUE, updated=$1 WHERE id=$2`, now, id)
	if err != nil {
		return nil, err
	}
	q.Hidden = true
	q.Updated = now
	return q, nil
}

func (p *Postgres) QuoteAddLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error) {
	q, err := p.getQuote(ctx, id)
	if err != nil {
		return nil, err
	}
	if slices.Contains(q.LineItemIds, lineItemId) {
		return q, nil
	}
	q.LineItemIds = append(q.LineItemIds, lineItemId)
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE quotes SET line_item_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(q.LineItemIds), now, id)
	if err != nil {
		return nil, err
	}
	q.Updated = now
	return q, nil
}

func (p *Postgres) QuoteRemoveLineItem(ctx context.Context, id ID, lineItemId ID) (*Quote, error) {
	q, err := p.getQuote(ctx, id)
	if err != nil {
		return nil, err
	}
	for i, li := range q.LineItemIds {
		if li == lineItemId {
			q.LineItemIds = slices.Delete(q.LineItemIds, i, i+1)
			break
		}
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE quotes SET line_item_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(q.LineItemIds), now, id)
	if err != nil {
		return nil, err
	}
	q.Updated = now
	return q, nil
}

func (p *Postgres) QuoteAddAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error) {
	q, err := p.getQuote(ctx, id)
	if err != nil {
		return nil, err
	}
	if slices.Contains(q.AdjustmentIds, adjustmentId) {
		return q, nil
	}
	q.AdjustmentIds = append(q.AdjustmentIds, adjustmentId)
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE quotes SET adjustment_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(q.AdjustmentIds), now, id)
	if err != nil {
		return nil, err
	}
	q.Updated = now
	return q, nil
}

func (p *Postgres) QuoteRemoveAdjustment(ctx context.Context, id ID, adjustmentId ID) (*Quote, error) {
	q, err := p.getQuote(ctx, id)
	if err != nil {
		return nil, err
	}
	for i, a := range q.AdjustmentIds {
		if a == adjustmentId {
			q.AdjustmentIds = slices.Delete(q.AdjustmentIds, i, i+1)
			break
		}
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE quotes SET adjustment_ids=$1, updated=$2 WHERE id=$3`,
		mustMarshal(q.AdjustmentIds), now, id)
	if err != nil {
		return nil, err
	}
	q.Updated = now
	return q, nil
}

// ─────────────────────────────────────────────
// LINE ITEM
// ─────────────────────────────────────────────

const lineItemCols = `id, quote_id, parent_id, sub_item_ids, image_id, description, quantity, quantity_suffix, quantity_prefix, unit_price, unit_price_suffix, unit_price_prefix, amount, amount_suffix, amount_prefix, open, created, updated`

func (p *Postgres) createLineItemRaw(ctx context.Context, quoteId ID, parentId *ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO line_items (id, org_id, group_id, quote_id, parent_id, sub_item_ids, image_id, description, quantity, quantity_suffix, quantity_prefix, unit_price, unit_price_suffix, unit_price_prefix, amount, amount_suffix, amount_prefix, open, created, updated)
		VALUES ($1,$2,$3,$4,$5,'[]',NULL,$6,$7,'','',$8,'','', $9,'','',FALSE,$10,$11)`,
		id, orgId, groupId, quoteId, parentId, description, quantity, unitPrice, amount, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &LineItem{
		Id:          id,
		QuoteId:     quoteId,
		ParentId:    parentId,
		SubItemIds:  []ID{},
		Description: description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		Amount:      amount,
		Created:     now,
		Updated:     now,
	}, nil
}

func (p *Postgres) CreateLineItem(ctx context.Context, quoteId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error) {
	return p.createLineItemRaw(ctx, quoteId, nil, description, quantity, unitPrice, amount)
}

func (p *Postgres) CreateSubLineItem(ctx context.Context, quoteId, parentId ID, description string, quantity, unitPrice int, amount *int) (*LineItem, error) {
	return p.createLineItemRaw(ctx, quoteId, &parentId, description, quantity, unitPrice, amount)
}

func (p *Postgres) CreateDuplicateLineItem(ctx context.Context, id ID) (*LineItem, error) {
	orig, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return p.createLineItemRaw(ctx, orig.QuoteId, orig.ParentId, orig.Description, orig.Quantity, orig.UnitPrice, orig.Amount)
}

func (p *Postgres) GetLineItem(ctx context.Context, id ID) (*LineItem, error) {
	rows, err := p.db.Query(ctx, `SELECT `+lineItemCols+` FROM line_items WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	li, err := pgx.CollectOneRow(rows, pgxRowToLineItem)
	if err != nil {
		return nil, err
	}
	var orgId, groupId string
	err = p.db.QueryRow(ctx, `SELECT org_id, group_id FROM line_items WHERE id=$1`, id).Scan(&orgId, &groupId)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, orgId, groupId); err != nil {
		return nil, err
	}
	return li, nil
}

func (p *Postgres) MoveLineItem(ctx context.Context, id ID, parentId *ID, index *int) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET parent_id=$1, updated=$2 WHERE id=$3`,
		parentId, now, id)
	if err != nil {
		return nil, err
	}
	li.ParentId = parentId
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemImage(ctx context.Context, id ID, imageId *ID) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET image_id=$1, updated=$2 WHERE id=$3`,
		imageId, now, id)
	if err != nil {
		return nil, err
	}
	li.ImageId = imageId
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemDescription(ctx context.Context, id ID, description string) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET description=$1, updated=$2 WHERE id=$3`,
		description, now, id)
	if err != nil {
		return nil, err
	}
	li.Description = description
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemQuantity(ctx context.Context, id ID, quantity int, prefix, suffix string) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET quantity=$1, quantity_prefix=$2, quantity_suffix=$3, updated=$4 WHERE id=$5`,
		quantity, prefix, suffix, now, id)
	if err != nil {
		return nil, err
	}
	li.Quantity = quantity
	li.QuantityPrefix = prefix
	li.QuantitySuffix = suffix
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemUnitPrice(ctx context.Context, id ID, unitPrice int, prefix, suffix string) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET unit_price=$1, unit_price_prefix=$2, unit_price_suffix=$3, updated=$4 WHERE id=$5`,
		unitPrice, prefix, suffix, now, id)
	if err != nil {
		return nil, err
	}
	li.UnitPrice = unitPrice
	li.UnitPricePrefix = prefix
	li.UnitPriceSuffix = suffix
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemAmount(ctx context.Context, id ID, amount *int, prefix, suffix string) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET amount=$1, amount_prefix=$2, amount_suffix=$3, updated=$4 WHERE id=$5`,
		amount, prefix, suffix, now, id)
	if err != nil {
		return nil, err
	}
	li.Amount = amount
	li.AmountPrefix = prefix
	li.AmountSuffix = suffix
	li.Updated = now
	return li, nil
}

func (p *Postgres) UpdateLineItemOpen(ctx context.Context, id ID, open bool) (*LineItem, error) {
	li, err := p.GetLineItem(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE line_items SET open=$1, updated=$2 WHERE id=$3`, open, now, id)
	if err != nil {
		return nil, err
	}
	li.Open = open
	li.Updated = now
	return li, nil
}

func (p *Postgres) DeleteLineItem(ctx context.Context, id ID) error {
	if _, err := p.GetLineItem(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `DELETE FROM line_items WHERE id=$1`, id)
	return err
}

// ─────────────────────────────────────────────
// ADJUSTMENT
// ─────────────────────────────────────────────

const adjCols = `id, quote_id, description, type, amount, created, updated`

func (p *Postgres) CreateAdjustment(ctx context.Context, quoteId ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error) {
	orgId, groupId, err := p.ext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	id := newID()
	_, err = p.db.Exec(ctx, `
		INSERT INTO adjustments (id, org_id, group_id, quote_id, description, type, amount, created, updated)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, orgId, groupId, quoteId, description, adjustmentType, amount, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &Adjustment{
		Id: id, QuoteId: quoteId, Description: description,
		Type: adjustmentType, Amount: amount,
		Created: now, Updated: now,
	}, nil
}

func (p *Postgres) GetAdjustment(ctx context.Context, id ID) (*Adjustment, error) {
	rows, err := p.db.Query(ctx, `SELECT `+adjCols+` FROM adjustments WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	a, err := pgx.CollectOneRow(rows, pgxRowToAdjustment)
	if err != nil {
		return nil, err
	}
	var orgId, groupId string
	err = p.db.QueryRow(ctx, `SELECT org_id, group_id FROM adjustments WHERE id=$1`, id).Scan(&orgId, &groupId)
	if err != nil {
		return nil, err
	}
	if err := p.authCheck(ctx, orgId, groupId); err != nil {
		return nil, err
	}
	return a, nil
}

func (p *Postgres) UpdateAdjustment(ctx context.Context, id ID, description string, amount int, adjustmentType AdjustmentType) (*Adjustment, error) {
	a, err := p.GetAdjustment(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = p.db.Exec(ctx, `UPDATE adjustments SET description=$1, amount=$2, type=$3, updated=$4 WHERE id=$5`,
		description, amount, adjustmentType, now, id)
	if err != nil {
		return nil, err
	}
	a.Description = description
	a.Amount = amount
	a.Type = adjustmentType
	a.Updated = now
	return a, nil
}

func (p *Postgres) RemoveAdjustment(ctx context.Context, id ID) error {
	if _, err := p.GetAdjustment(ctx, id); err != nil {
		return err
	}
	_, err := p.db.Exec(ctx, `DELETE FROM adjustments WHERE id=$1`, id)
	return err
}

// ─────────────────────────────────────────────
// CONTACT
// ─────────────────────────────────────────────

func (p *Postgres) GetContact(ctx context.Context, id ID) (*Contact, error) {
	rows, err := p.db.Query(ctx,
		`SELECT id, name, company_name, phones, emails, websites, street, city, state, zip FROM contacts WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(rows, pgxRowToContact)
}

// ─────────────────────────────────────────────
// TRANSACTION
// ─────────────────────────────────────────────

func (p *Postgres) Transaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	// We inject the transaction into the context so that operations within f
	// can participate. However, the existing Store methods use p.db directly,
	// so we wrap them via a txStore shim.
	txStore := &pgTxStore{Postgres: p, tx: tx}
	txCtx := context.WithValue(ctx, pgTxKey{}, txStore)
	if err := f(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// pgTxKey is used to store a transaction-aware store in the context.
type pgTxKey struct{}

// pgTxStore wraps Postgres but executes queries against a pgx.Tx.
// It embeds *Postgres and overrides the db field by shadowing Exec/Query calls.
// For simplicity, we use a db interface abstraction.
type pgTxStore struct {
	*Postgres
	tx pgx.Tx
}

// ─────────────────────────────────────────────
// COMPILE-TIME INTERFACE CHECK
// ─────────────────────────────────────────────

var _ Store = (*Postgres)(nil)

// ErrNotFound is returned when a record is not found.
var ErrNotFound = errors.New("record not found")
