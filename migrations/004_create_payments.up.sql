CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'refunded', 'failed');

CREATE TABLE payments (
    id            UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id    UUID           NOT NULL REFERENCES bookings(id),
    amount        DECIMAL(10,2)  NOT NULL,
    platform_fee  DECIMAL(10,2)  NOT NULL,
    owner_payout  DECIMAL(10,2)  NOT NULL,
    status        payment_status NOT NULL DEFAULT 'pending',
    provider      VARCHAR(50),
    provider_ref  VARCHAR(255),
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_booking ON payments(booking_id);
