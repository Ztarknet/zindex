-- Migration: Add input_count, output_count, and balance_change fields
-- Description: Adds new fields to existing tables and backfills data
-- Date: 2025-01-20

BEGIN;

-- ============================================================================
-- STEP 1: Add new columns to transactions table
-- ============================================================================

DO $$
BEGIN
    -- Add input_count column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'input_count'
    ) THEN
        ALTER TABLE transactions ADD COLUMN input_count INT NOT NULL DEFAULT 0;
        RAISE NOTICE 'Added input_count column to transactions table';
    ELSE
        RAISE NOTICE 'input_count column already exists in transactions table';
    END IF;

    -- Add output_count column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'output_count'
    ) THEN
        ALTER TABLE transactions ADD COLUMN output_count INT NOT NULL DEFAULT 0;
        RAISE NOTICE 'Added output_count column to transactions table';
    ELSE
        RAISE NOTICE 'output_count column already exists in transactions table';
    END IF;
END $$;

-- ============================================================================
-- STEP 2: Backfill input_count and output_count for existing transactions
-- ============================================================================

-- Update input_count based on transaction_inputs
UPDATE transactions t
SET input_count = COALESCE(
    (SELECT COUNT(*) FROM transaction_inputs ti WHERE ti.txid = t.txid),
    0
)
WHERE input_count = 0;

-- Update output_count based on transaction_outputs
UPDATE transactions t
SET output_count = COALESCE(
    (SELECT COUNT(*) FROM transaction_outputs tvo WHERE tvo.txid = t.txid),
    0
)
WHERE output_count = 0;

RAISE NOTICE 'Backfilled input_count and output_count for existing transactions';

-- ============================================================================
-- STEP 3: Add balance_change column to account_transactions table
-- ============================================================================

DO $$
BEGIN
    -- Add balance_change column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'account_transactions' AND column_name = 'balance_change'
    ) THEN
        ALTER TABLE account_transactions ADD COLUMN balance_change BIGINT NOT NULL DEFAULT 0;
        RAISE NOTICE 'Added balance_change column to account_transactions table';
    ELSE
        RAISE NOTICE 'balance_change column already exists in account_transactions table';
    END IF;
END $$;

-- ============================================================================
-- STEP 4: Backfill balance_change for existing account_transactions
-- ============================================================================

-- For receiving transactions: sum all outputs going to this address in this transaction
UPDATE account_transactions at
SET balance_change = COALESCE(
    (
        SELECT COALESCE(SUM(tvo.value), 0)
        FROM transaction_outputs tvo
        WHERE tvo.txid = at.txid
        -- Note: We can't directly match outputs to addresses without scriptPubKey info
        -- This is a limitation - we'll set it based on transaction type
    ),
    0
)
WHERE at.type = 'receive' AND balance_change = 0;

-- For sending transactions: negative sum of inputs (approximation)
-- Note: This is an approximation since we don't have perfect input-to-address mapping
UPDATE account_transactions at
SET balance_change = COALESCE(
    (
        SELECT -COALESCE(SUM(ti.value), 0)
        FROM transaction_inputs ti
        WHERE ti.txid = at.txid
    ),
    0
)
WHERE at.type = 'send' AND balance_change = 0;

RAISE NOTICE 'Backfilled balance_change for existing account_transactions';

-- ============================================================================
-- STEP 5: Verify the migration
-- ============================================================================

-- Check transactions table
DO $$
DECLARE
    tx_count INT;
    tx_with_counts INT;
BEGIN
    SELECT COUNT(*) INTO tx_count FROM transactions;
    SELECT COUNT(*) INTO tx_with_counts FROM transactions WHERE input_count > 0 OR output_count > 0;

    RAISE NOTICE 'Transactions table: % total, % with counts populated', tx_count, tx_with_counts;
END $$;

-- Check account_transactions table
DO $$
DECLARE
    at_count INT;
    at_with_balance INT;
BEGIN
    SELECT COUNT(*) INTO at_count FROM account_transactions;
    SELECT COUNT(*) INTO at_with_balance FROM account_transactions WHERE balance_change != 0;

    RAISE NOTICE 'Account_transactions table: % total, % with balance_change populated', at_count, at_with_balance;
END $$;

COMMIT;

-- ============================================================================
-- Migration completed successfully
-- ============================================================================
