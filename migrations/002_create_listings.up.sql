CREATE TYPE listing_category AS ENUM (
    'excavator', 'crane', 'scaffold', 'compactor', 'loader', 'forklift', 'generator', 'other'
);

CREATE TABLE listings (
    id           UUID             PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id     UUID             NOT NULL REFERENCES users(id),
    title        VARCHAR(200)     NOT NULL,
    category     listing_category NOT NULL,
    description  TEXT             NOT NULL,
    location     VARCHAR(255)     NOT NULL,
    price_per_day DECIMAL(10,2)   NOT NULL CHECK (price_per_day > 0),
    minimum_days INT              NOT NULL DEFAULT 1,
    images       TEXT[]           NOT NULL DEFAULT '{}',
    specs        JSONB            NOT NULL DEFAULT '{}',
    is_available BOOLEAN          NOT NULL DEFAULT true,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_listings_owner    ON listings(owner_id);
CREATE INDEX idx_listings_category ON listings(category);
CREATE INDEX idx_listings_location ON listings(location);
CREATE INDEX idx_listings_active   ON listings(is_available) WHERE deleted_at IS NULL;
