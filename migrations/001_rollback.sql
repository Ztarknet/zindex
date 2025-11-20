-- Rollback Migration: Remove input_count, output_count, and balance_change fields
-- Description: Removes the fields added in migration 001
-- Date: 2025-01-20
-- WARNING: This will drop the columns and lose the data in them

BEGIN;

-- ============================================================================
-- Remove columns from transactions table
-- ============================================================================

DO $$
BEGIN
    -- Remove input_count column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'input_count'
    ) THEN
        ALTER TABLE transactions DROP COLUMN input_count;
        RAISE NOTICE '✓ Removed input_count column from transactions table';
    ELSE
        RAISE NOTICE '- input_count column does not exist in transactions table';
    END IF;

    -- Remove output_count column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'transactions' AND column_name = 'output_count'
    ) THEN
        ALTER TABLE transactions DROP COLUMN output_count;
        RAISE NOTICE '✓ Removed output_count column from transactions table';
    ELSE
        RAISE NOTICE '- output_count column does not exist in transactions table';
    END IF;
END $$;

-- ============================================================================
-- Remove column from account_transactions table
-- ============================================================================

DO $$
BEGIN
    -- Remove balance_change column if it exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'account_transactions' AND column_name = 'balance_change'
    ) THEN
        ALTER TABLE account_transactions DROP COLUMN balance_change;
        RAISE NOTICE '✓ Removed balance_change column from account_transactions table';
    ELSE
        RAISE NOTICE '- balance_change column does not exist in account_transactions table';
    END IF;
END $$;

COMMIT;

RAISE NOTICE '';
RAISE NOTICE '========================================';
RAISE NOTICE 'Rollback completed successfully!';
RAISE NOTICE '========================================';
