#!/bin/bash
# Revert zindex to a specific block height
# Usage: ./revert-to.sh <block_height>
# Example: ./revert-to.sh 1330

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check arguments
if [ -z "$1" ]; then
    echo -e "${RED}Error: Block height required${NC}"
    echo "Usage: ./revert-to.sh <block_height>"
    echo "Example: ./revert-to.sh 1330"
    exit 1
fi

REVERT_HEIGHT=$1
NAMESPACE="${NAMESPACE:-default}"

# Validate block height is a number
if ! [[ "$REVERT_HEIGHT" =~ ^[0-9]+$ ]]; then
    echo -e "${RED}Error: Block height must be a positive integer${NC}"
    exit 1
fi

echo -e "${YELLOW}=== Zindex Revert Script ===${NC}"
echo -e "Reverting to block height: ${GREEN}$REVERT_HEIGHT${NC}"
echo -e "Namespace: ${GREEN}$NAMESPACE${NC}"
echo ""

# Step 1: Find deployments/pods
echo -e "${YELLOW}[1/5] Finding Kubernetes resources...${NC}"

# Find zindex deployment
ZINDEX_DEPLOYMENT=$(kubectl get deployments -n "$NAMESPACE" -o name 2>/dev/null | grep -E "zindex" | head -1)
if [ -z "$ZINDEX_DEPLOYMENT" ]; then
    echo -e "${RED}Error: Could not find zindex deployment in namespace $NAMESPACE${NC}"
    echo "Available deployments:"
    kubectl get deployments -n "$NAMESPACE" -o name
    exit 1
fi
echo -e "  Found zindex deployment: ${GREEN}$ZINDEX_DEPLOYMENT${NC}"

# Find postgres pod
POSTGRES_POD=$(kubectl get pods -n "$NAMESPACE" -o name 2>/dev/null | grep -E "postgres" | head -1)
if [ -z "$POSTGRES_POD" ]; then
    echo -e "${RED}Error: Could not find postgres pod in namespace $NAMESPACE${NC}"
    echo "Available pods:"
    kubectl get pods -n "$NAMESPACE" -o name
    exit 1
fi
# Strip "pod/" prefix
POSTGRES_POD=${POSTGRES_POD#pod/}
echo -e "  Found postgres pod: ${GREEN}$POSTGRES_POD${NC}"

# Get database credentials from environment or use defaults
DB_NAME="${DB_NAME:-zindex}"
DB_USER="${DB_USER:-zindex}"

echo -e "  Database: ${GREEN}$DB_NAME${NC}"
echo -e "  User: ${GREEN}$DB_USER${NC}"
echo ""

# Step 2: Scale down zindex
echo -e "${YELLOW}[2/5] Scaling down zindex to 0 replicas...${NC}"
kubectl scale "$ZINDEX_DEPLOYMENT" -n "$NAMESPACE" --replicas=0
echo -e "  ${GREEN}Scaled down${NC}"

# Wait for pods to terminate
echo -e "  Waiting for zindex pods to terminate..."
kubectl rollout status "$ZINDEX_DEPLOYMENT" -n "$NAMESPACE" --timeout=60s 2>/dev/null || true
sleep 2
echo ""

# Step 3: Get current state before revert
echo -e "${YELLOW}[3/5] Checking current database state...${NC}"

CURRENT_BLOCK=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- \
    psql -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COALESCE(last_indexed_block, 0) FROM indexer_state LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

if [ -z "$CURRENT_BLOCK" ] || [ "$CURRENT_BLOCK" = "" ]; then
    CURRENT_BLOCK=0
fi

echo -e "  Current last indexed block: ${GREEN}$CURRENT_BLOCK${NC}"

if [ "$REVERT_HEIGHT" -ge "$CURRENT_BLOCK" ]; then
    echo -e "${YELLOW}Warning: Revert height ($REVERT_HEIGHT) >= current height ($CURRENT_BLOCK)${NC}"
    echo -e "${YELLOW}Nothing to revert. Scaling zindex back up...${NC}"
    kubectl scale "$ZINDEX_DEPLOYMENT" -n "$NAMESPACE" --replicas=1
    exit 0
fi

# Step 4: Execute rollback SQL
echo -e "${YELLOW}[4/5] Executing rollback queries...${NC}"

# Build SQL script
SQL_SCRIPT=$(cat <<EOF
-- Zindex Rollback to block $REVERT_HEIGHT
-- Generated at $(date -u +"%Y-%m-%d %H:%M:%S UTC")

BEGIN;

-- 1. Unspend transaction outputs that were spent at height > $REVERT_HEIGHT
UPDATE transaction_outputs
SET spent_by_txid = NULL, spent_by_vin = NULL, spent_at_height = NULL
WHERE spent_at_height > $REVERT_HEIGHT;

-- 2. Unspend TZE outputs that were spent at height > $REVERT_HEIGHT
UPDATE tze_outputs
SET spent_by_txid = NULL, spent_by_vin = NULL, spent_at_height = NULL
WHERE spent_at_height > $REVERT_HEIGHT;

-- 3. Delete account_transactions for blocks > $REVERT_HEIGHT
DELETE FROM account_transactions WHERE block_height > $REVERT_HEIGHT;

-- 4. Recalculate account balances from remaining transactions
UPDATE accounts a
SET balance = COALESCE((
    SELECT SUM(balance_change)
    FROM account_transactions at
    WHERE at.address = a.address
), 0);

-- 5. Delete orphaned accounts (no transactions left)
DELETE FROM accounts
WHERE NOT EXISTS (
    SELECT 1 FROM account_transactions WHERE address = accounts.address
);

-- 6. Delete TZE inputs for transactions in blocks > $REVERT_HEIGHT
DELETE FROM tze_inputs
WHERE txid IN (SELECT txid FROM transactions WHERE block_height > $REVERT_HEIGHT);

-- 7. Delete TZE outputs for transactions in blocks > $REVERT_HEIGHT
DELETE FROM tze_outputs
WHERE txid IN (SELECT txid FROM transactions WHERE block_height > $REVERT_HEIGHT);

-- 8. Delete STARK proofs for blocks > $REVERT_HEIGHT
DELETE FROM stark_proofs WHERE block_height > $REVERT_HEIGHT;

-- 9. Delete Ztarknet facts for blocks > $REVERT_HEIGHT
DELETE FROM ztarknet_facts WHERE block_height > $REVERT_HEIGHT;

-- 10. Delete orphaned verifiers (no proofs or facts left)
DELETE FROM verifiers
WHERE NOT EXISTS (
    SELECT 1 FROM stark_proofs WHERE verifier_id = verifiers.verifier_id
) AND NOT EXISTS (
    SELECT 1 FROM ztarknet_facts WHERE verifier_id = verifiers.verifier_id
);

-- 11. Delete transactions for blocks > $REVERT_HEIGHT (CASCADE deletes inputs/outputs)
DELETE FROM transactions WHERE block_height > $REVERT_HEIGHT;

-- 12. Delete blocks > $REVERT_HEIGHT
DELETE FROM blocks WHERE height > $REVERT_HEIGHT;

-- 13. Update indexer_state with new last indexed block
UPDATE indexer_state
SET last_indexed_block = $REVERT_HEIGHT,
    last_indexed_hash = (SELECT hash FROM blocks WHERE height = $REVERT_HEIGHT),
    updated_at = NOW();

-- If no rows updated, insert the state
INSERT INTO indexer_state (last_indexed_block, last_indexed_hash, updated_at)
SELECT $REVERT_HEIGHT,
       (SELECT hash FROM blocks WHERE height = $REVERT_HEIGHT),
       NOW()
WHERE NOT EXISTS (SELECT 1 FROM indexer_state);

COMMIT;

-- Report final state
SELECT 'Rollback complete. New last indexed block: ' || COALESCE(last_indexed_block::text, 'NULL') as status
FROM indexer_state LIMIT 1;
EOF
)

# Execute SQL
echo -e "  Executing SQL rollback..."
RESULT=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- \
    psql -U "$DB_USER" -d "$DB_NAME" -c "$SQL_SCRIPT" 2>&1)

echo "$RESULT"
echo ""

# Verify rollback
echo -e "${YELLOW}[5/5] Verifying rollback and scaling up zindex...${NC}"

NEW_BLOCK=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- \
    psql -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COALESCE(last_indexed_block, 0) FROM indexer_state LIMIT 1;" 2>/dev/null | tr -d '[:space:]')

BLOCK_COUNT=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- \
    psql -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM blocks;" 2>/dev/null | tr -d '[:space:]')

TX_COUNT=$(kubectl exec -n "$NAMESPACE" "$POSTGRES_POD" -- \
    psql -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT COUNT(*) FROM transactions;" 2>/dev/null | tr -d '[:space:]')

echo -e "  New last indexed block: ${GREEN}$NEW_BLOCK${NC}"
echo -e "  Total blocks in DB: ${GREEN}$BLOCK_COUNT${NC}"
echo -e "  Total transactions in DB: ${GREEN}$TX_COUNT${NC}"
echo ""

# Scale zindex back up
echo -e "  Scaling zindex back to 1 replica..."
kubectl scale "$ZINDEX_DEPLOYMENT" -n "$NAMESPACE" --replicas=1
kubectl rollout status "$ZINDEX_DEPLOYMENT" -n "$NAMESPACE" --timeout=120s

echo ""
echo -e "${GREEN}=== Rollback Complete ===${NC}"
echo -e "Reverted from block ${RED}$CURRENT_BLOCK${NC} to block ${GREEN}$NEW_BLOCK${NC}"
echo -e "Zindex will resume indexing from block ${GREEN}$((NEW_BLOCK + 1))${NC}"
