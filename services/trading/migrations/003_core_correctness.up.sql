ALTER TABLE accounts
    ADD CONSTRAINT chk_accounts_balance_non_negative
    CHECK (balance >= 0);

ALTER TABLE orders
    ADD CONSTRAINT chk_orders_quantity_positive
    CHECK (quantity > 0);

ALTER TABLE orders
    ADD CONSTRAINT chk_orders_price_positive
    CHECK (price > 0);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_quantity_positive
    CHECK (quantity > 0);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_entry_price_positive
    CHECK (entry_price > 0);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_mark_price_positive
    CHECK (mark_price > 0);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_initial_margin_non_negative
    CHECK (initial_margin >= 0);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_sl_close_percent_range
    CHECK (sl_close_percent BETWEEN 1 AND 100);

ALTER TABLE positions
    ADD CONSTRAINT chk_positions_tp_close_percent_range
    CHECK (tp_close_percent BETWEEN 1 AND 100);

ALTER TABLE trades
    ADD CONSTRAINT chk_trades_quantity_positive
    CHECK (quantity > 0);

ALTER TABLE trades
    ADD CONSTRAINT chk_trades_price_positive
    CHECK (price > 0);

CREATE TABLE idempotency_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    request_hash CHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('PROCESSING', 'COMPLETED', 'FAILED')),
    response_code INTEGER,
    response_body TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, scope, key)
);

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);
CREATE INDEX idx_idempotency_keys_user_created_at ON idempotency_keys(user_id, created_at DESC);

CREATE TRIGGER update_idempotency_keys_updated_at
    BEFORE UPDATE ON idempotency_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
