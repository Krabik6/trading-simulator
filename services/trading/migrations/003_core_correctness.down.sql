DROP TRIGGER IF EXISTS update_idempotency_keys_updated_at ON idempotency_keys;
DROP TABLE IF EXISTS idempotency_keys;

ALTER TABLE trades
    DROP CONSTRAINT IF EXISTS chk_trades_price_positive;

ALTER TABLE trades
    DROP CONSTRAINT IF EXISTS chk_trades_quantity_positive;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_tp_close_percent_range;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_sl_close_percent_range;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_initial_margin_non_negative;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_mark_price_positive;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_entry_price_positive;

ALTER TABLE positions
    DROP CONSTRAINT IF EXISTS chk_positions_quantity_positive;

ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS chk_orders_price_positive;

ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS chk_orders_quantity_positive;

ALTER TABLE accounts
    DROP CONSTRAINT IF EXISTS chk_accounts_balance_non_negative;
