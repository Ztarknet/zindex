# Zindex Database Migrations

This directory contains database migration scripts for zindex.

## Migration 001: Add Count and Balance Fields

This migration adds the following enhancements:

### Changes
1. **Transactions Table**: Adds `input_count` and `output_count` columns
2. **Account Transactions Table**: Adds `balance_change` column

### Files
- `001_add_count_and_balance_fields_improved.sql` - Main migration (recommended)
- `001_add_count_and_balance_fields.sql` - Alternative version
- `001_rollback.sql` - Rollback migration

---

## Prerequisites

1. **PostgreSQL Client**: Ensure `psql` is installed
2. **Database Access**: Have your database credentials ready
3. **Backup** (Recommended): Always backup before running migrations

---

## Quick Start

### Option 1: Using the Migration Runner (Recommended)

```bash
# Set database password
export DB_PASSWORD="your_password"

# Run migration with backup
cd migrations
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -b

# Run migration without backup (not recommended for production)
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -y
```

### Option 2: Manual Execution

```bash
# Create a backup first
pg_dump -h localhost -p 5432 -U postgres -d zindex > backup_$(date +%Y%m%d_%H%M%S).sql

# Run the migration
psql -h localhost -p 5432 -U postgres -d zindex -f 001_add_count_and_balance_fields_improved.sql
```

---

## Environment Variables

Configure these environment variables for the migration runner:

```bash
export ZINDEX_DB_HOST="localhost"      # Default: localhost
export ZINDEX_DB_PORT="5432"           # Default: 5432
export ZINDEX_DB_NAME="zindex"         # Default: zindex
export ZINDEX_DB_USER="postgres"       # Default: postgres
export DB_PASSWORD="your_password"     # Required
```

---

## Migration Details

### Step 1: Add Columns
- Checks if columns already exist before adding
- Uses `ALTER TABLE` to add columns to existing tables
- Sets default values (0) for new columns

### Step 2: Backfill Data

#### Input Count & Output Count
- Counts actual inputs from `transaction_inputs` table
- Counts actual outputs from `transaction_outputs` table
- Updates all existing transactions

#### Balance Change (Approximation for Old Data)
- **Receiving transactions**: Uses `total_output` from transaction
- **Sending transactions**: Uses negative of `(total_output + total_fee)`

> **Note**: Balance change values for existing data are approximations. New transactions indexed after the migration will have accurate values based on actual output values.

### Step 3: Verification
- Displays counts of affected rows
- Shows statistics for validation

---

## Testing the Migration

### On a Test Database First (Highly Recommended)

```bash
# 1. Create a test database with a copy of your data
createdb zindex_test
pg_dump zindex | psql zindex_test

# 2. Run migration on test database
export ZINDEX_DB_NAME="zindex_test"
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql

# 3. Verify the results
psql zindex_test -c "SELECT COUNT(*),
    COUNT(*) FILTER (WHERE input_count > 0) as with_inputs,
    COUNT(*) FILTER (WHERE output_count > 0) as with_outputs
    FROM transactions;"

# 4. If successful, run on production
```

---

## Rollback

If you need to rollback the migration:

```bash
# With backup
./run-migration.sh -f 001_rollback.sql -b

# Without backup (destructive!)
./run-migration.sh -f 001_rollback.sql -y
```

> **Warning**: Rollback will permanently delete the added columns and their data!

---

## Post-Migration Steps

### 1. Restart Zindex
After running the migration, restart your zindex service to ensure it uses the new schema:

```bash
# Stop zindex
systemctl stop zindex  # or your service manager

# Start zindex
systemctl start zindex
```

### 2. Verify New Data
Check that new transactions being indexed have proper values:

```bash
# Get a recent transaction
curl "http://localhost:8080/api/v1/tx-graph/transactions/recent?limit=1" | jq

# Should show input_count and output_count fields
```

### 3. Monitor Logs
Watch zindex logs for any issues:

```bash
journalctl -u zindex -f  # or your logging system
```

---

## Troubleshooting

### Migration Fails

1. **Check database connection**:
   ```bash
   psql -h $ZINDEX_DB_HOST -p $ZINDEX_DB_PORT -U $ZINDEX_DB_USER -d $ZINDEX_DB_NAME -c "SELECT 1;"
   ```

2. **Check for locks**:
   ```sql
   SELECT * FROM pg_locks WHERE NOT granted;
   ```

3. **Review error messages**: The migration uses transactions, so failures will rollback automatically

### Columns Already Exist

If columns already exist, the migration will skip adding them and only backfill data.

### Performance Concerns

For very large databases:
- Consider running during low-traffic periods
- The backfill queries use efficient aggregations
- Expected time: ~1-10 seconds per million transactions

### Backup Issues

If `pg_dump` fails:
- Check disk space
- Verify user permissions
- Try backing up specific tables:
  ```bash
  pg_dump -t transactions -t account_transactions zindex > backup.sql
  ```

---

## Migration Runner Options

```
Usage: ./run-migration.sh [OPTIONS]

Options:
  -f, --file FILE       Migration file to run (required)
  -b, --backup          Create backup before migration
  -y, --yes             Skip confirmation prompt
  -h, --help            Show help message

Examples:
  ./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql
  ./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -b
  ./run-migration.sh -f 001_rollback.sql -b -y
```

---

## Best Practices

1. ✅ **Always test on a copy first**
2. ✅ **Create a backup before migrating production**
3. ✅ **Run during low-traffic periods**
4. ✅ **Monitor the migration progress**
5. ✅ **Verify data after migration**
6. ✅ **Keep the backup until you're confident**

---

## Support

If you encounter issues:
1. Check the troubleshooting section above
2. Review zindex logs
3. Verify your database connection settings
4. Ensure you have sufficient permissions (ALTER TABLE, SELECT, UPDATE)

---

## Files in This Directory

```
migrations/
├── README.md                                          # This file
├── run-migration.sh                                   # Migration runner script
├── 001_add_count_and_balance_fields.sql              # Original migration
├── 001_add_count_and_balance_fields_improved.sql     # Improved migration (recommended)
└── 001_rollback.sql                                  # Rollback migration
```
