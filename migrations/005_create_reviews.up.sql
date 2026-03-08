CREATE TABLE reviews (
    id           UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id   UUID        NOT NULL REFERENCES bookings(id),
    reviewer_id  UUID        NOT NULL REFERENCES users(id),
    reviewee_id  UUID        REFERENCES users(id),     -- null when reviewing the listing itself
    listing_id   UUID        REFERENCES listings(id),  -- null when reviewing a renter
    rating       INT         NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment      TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (booking_id, reviewer_id)                   -- one review per party per booking
);

CREATE INDEX idx_reviews_listing  ON reviews(listing_id);
CREATE INDEX idx_reviews_reviewee ON reviews(reviewee_id);
