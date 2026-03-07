-- Pricey library schema
-- All statements use CREATE TABLE IF NOT EXISTS; safe to run on every startup.

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
