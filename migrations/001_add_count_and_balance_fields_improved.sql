-- Migration: Add input_count, output_count, and balance_change fields (Improved)
-- Description: Adds new fields to existing tables and backfills data accurately
-- Date: 2025-01-20
-- Note: This migration handles balance_change more carefully

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
        RAISE NOTICE '✓ Added input_count column to transactions table';
    ELSE
        RAISE NOTICE '- input_count column already exists in transactions table';
    END IF;

    -- Add output_count column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'output_count'
    ) THEN
        ALTER TABLE transactions ADD COLUMN output_count INT NOT NULL DEFAULT 0;
        RAISE NOTICE '✓ Added output_count column to transactions table';
    ELSE
        RAISE NOTICE '- output_count column already exists in transactions table';
    END IF;
END $$;

-- ============================================================================
-- STEP 2: Backfill input_count and output_count for existing transactions
-- ============================================================================

DO $$
DECLARE
    updated_rows INT;
BEGIN
    -- Update input_count based on transaction_inputs
    WITH input_counts AS (
        SELECT txid, COUNT(*) as cnt
        FROM transaction_inputs
        GROUP BY txid
    )
    UPDATE transactions t
    SET input_count = ic.cnt
    FROM input_counts ic
    WHERE t.txid = ic.txid AND t.input_count = 0;

    GET DIAGNOSTICS updated_rows = ROW_COUNT;
    RAISE NOTICE '✓ Updated input_count for % transactions', updated_rows;

    -- Update output_count based on transaction_outputs
    WITH output_counts AS (
        SELECT txid, COUNT(*) as cnt
        FROM transaction_outputs
        GROUP BY txid
    )
    UPDATE transactions t
    SET output_count = oc.cnt
    FROM output_counts oc
    WHERE t.txid = oc.txid AND t.output_count = 0;

    GET DIAGNOSTICS updated_rows = ROW_COUNT;
    RAISE NOTICE '✓ Updated output_count for % transactions', updated_rows;
END $$;

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
        RAISE NOTICE '✓ Added balance_change column to account_transactions table';
    ELSE
        RAISE NOTICE '- balance_change column already exists in account_transactions table';
    END IF;
END $$;

-- ============================================================================
-- STEP 4: Backfill balance_change for existing account_transactions
-- ============================================================================
-- NOTE: This is a best-effort approach. The exact balance change per account
-- per transaction would require knowing which specific outputs belong to which
-- addresses. We use the transaction's total_output as an approximation for
-- receiving transactions.

DO $$
DECLARE
    updated_rows INT;
BEGIN
    -- For receiving transactions: use a portion of the transaction's total output
    -- This is an approximation - ideally we'd need output->address mapping
    UPDATE account_transactions at
    SET balance_change = (
        SELECT COALESCE(t.total_output, 0)
        FROM transactions t
        WHERE t.txid = at.txid
    )
    WHERE at.type = 'receive' AND at.balance_change = 0;

    GET DIAGNOSTICS updated_rows = ROW_COUNT;
    RAISE NOTICE '✓ Updated balance_change for % receiving account_transactions', updated_rows;

    -- For sending transactions: use negative of total output + fee
    -- This represents funds leaving the account
    UPDATE account_transactions at
    SET balance_change = -(
        SELECT COALESCE(t.total_output + t.total_fee, 0)
        FROM transactions t
        WHERE t.txid = at.txid
    )
    WHERE at.type = 'send' AND at.balance_change = 0;

    GET DIAGNOSTICS updated_rows = ROW_COUNT;
    RAISE NOTICE '✓ Updated balance_change for % sending account_transactions', updated_rows;

    RAISE NOTICE '';
    RAISE NOTICE 'NOTE: balance_change values are approximations for existing data.';
    RAISE NOTICE 'New transactions indexed after this migration will have accurate values.';
END $$;

-- ============================================================================
-- STEP 5: Verify the migration
-- ============================================================================

DO $$
DECLARE
    tx_count INT;
    tx_with_input_counts INT;
    tx_with_output_counts INT;
    at_count INT;
    at_with_balance INT;
    at_receive_count INT;
    at_send_count INT;
BEGIN
    -- Check transactions
    SELECT COUNT(*) INTO tx_count FROM transactions;
    SELECT COUNT(*) INTO tx_with_input_counts FROM transactions WHERE input_count > 0;
    SELECT COUNT(*) INTO tx_with_output_counts FROM transactions WHERE output_count > 0;

    RAISE NOTICE '';
    RAISE NOTICE '=== Transactions Table ===';
    RAISE NOTICE 'Total transactions: %', tx_count;
    RAISE NOTICE 'With input_count > 0: %', tx_with_input_counts;
    RAISE NOTICE 'With output_count > 0: %', tx_with_output_counts;

    -- Check account_transactions
    SELECT COUNT(*) INTO at_count FROM account_transactions;
    SELECT COUNT(*) INTO at_with_balance FROM account_transactions WHERE balance_change != 0;
    SELECT COUNT(*) INTO at_receive_count FROM account_transactions WHERE type = 'receive';
    SELECT COUNT(*) INTO at_send_count FROM account_transactions WHERE type = 'send';

    RAISE NOTICE '';
    RAISE NOTICE '=== Account Transactions Table ===';
    RAISE NOTICE 'Total account_transactions: %', at_count;
    RAISE NOTICE 'With balance_change != 0: %', at_with_balance;
    RAISE NOTICE 'Receiving transactions: %', at_receive_count;
    RAISE NOTICE 'Sending transactions: %', at_send_count;
    RAISE NOTICE '';
END $$;

COMMIT;

RAISE NOTICE '========================================';
RAISE NOTICE 'Migration completed successfully!';
RAISE NOTICE '========================================';
