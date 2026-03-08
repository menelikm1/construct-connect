CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'active', 'completed', 'cancelled');

CREATE TABLE bookings (
    id                  UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    listing_id          UUID           NOT NULL REFERENCES listings(id),
    renter_id           UUID           NOT NULL REFERENCES users(id),
    owner_id            UUID           NOT NULL REFERENCES users(id),
    start_date          DATE           NOT NULL,
    end_date            DATE           NOT NULL,
    total_days          INT            NOT NULL,
    total_price         DECIMAL(10,2)  NOT NULL,
    status              booking_status NOT NULL DEFAULT 'pending',
    cancellation_reason TEXT,
    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CHECK (end_date > start_date)
);

CREATE INDEX idx_bookings_listing ON bookings(listing_id);
CREATE INDEX idx_bookings_renter  ON bookings(renter_id);
CREATE INDEX idx_bookings_owner   ON bookings(owner_id);
CREATE INDEX idx_bookings_status  ON bookings(status);
