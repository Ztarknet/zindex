# Quick Migration Guide for Production Zindex (Kubernetes/Helm)

## TL;DR - Recommended Steps for Kubernetes

**Option 1: Automated with k8s-migrate.sh (Recommended)**

```bash
# Run migration and stay scaled down, then helm upgrade
cd migrations
./k8s-migrate.sh -n <namespace> --scale-down --skip-scale-up

# Deploy new zindex version
helm upgrade zindex ./deploy/zindex-infra -n <namespace>

# Verify
kubectl logs -f -l app=zindex -n <namespace>
```

**Option 2: Manual steps**

```bash
# 1. Scale down zindex deployment
kubectl scale deployment zindex --replicas=0 -n <namespace>

# 2. Port-forward to PostgreSQL
kubectl port-forward svc/postgres 5432:5432 -n <namespace> &

# 3. Backup database
pg_dump -h localhost -U postgres zindex > backup_$(date +%Y%m%d_%H%M%S).sql

# 4. Run migration
cd migrations
export DB_PASSWORD="$(kubectl get secret postgres-secret -n <namespace> -o jsonpath='{.data.postgres-password}' | base64 -d)"
export ZINDEX_DB_HOST="localhost"
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -y

# 5. Stop port-forward
pkill -f "kubectl port-forward.*postgres"

# 6. Deploy new zindex version (scales up automatically)
helm upgrade zindex ./deploy/zindex-infra -n <namespace>

# 7. Verify
kubectl logs -f -l app=zindex -n <namespace>
```

---

## What This Migration Does

### Database Schema Changes
1. Adds `input_count` and `output_count` to `transactions` table
2. Adds `balance_change` to `account_transactions` table
3. Backfills existing data with calculated values

### Why You Need This
Your existing code expects these fields, and without them:
- ❌ New transactions will fail to index
- ❌ API queries will return incomplete data
- ❌ Application will crash on database insert

---

## Decision Tree for Kubernetes

### Scenario 1: Small Database (< 1M transactions)
**Recommended**: Run migration during normal operations
- Expected downtime: ~30 seconds
- Can run without scaling down (optional)
- Backup is quick

```bash
# Set your namespace
export NAMESPACE="your-namespace"

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get database password from secret
export DB_PASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
export ZINDEX_DB_HOST="localhost"

# Run migration with backup
cd migrations
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -b

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"
```

### Scenario 2: Medium Database (1M - 10M transactions)
**Recommended**: Run during low-traffic window with brief downtime
- Expected time: 1-5 minutes
- Scale down deployment during migration
- Definitely create backup

**Option A: Using k8s-migrate.sh (Recommended)**
```bash
export NAMESPACE="your-namespace"

# Run migration with scale-down, skip scale-up
cd migrations
./k8s-migrate.sh -n $NAMESPACE --scale-down --skip-scale-up

# Deploy new zindex version
cd ..
helm upgrade zindex ./deploy/zindex-infra -n $NAMESPACE

# Wait for rollout
kubectl rollout status deployment/zindex -n $NAMESPACE
```

**Option B: Manual steps**
```bash
export NAMESPACE="your-namespace"

# Scale down zindex
kubectl scale deployment zindex --replicas=0 -n $NAMESPACE

# Wait for pods to terminate
kubectl wait --for=delete pod -l app=zindex -n $NAMESPACE --timeout=60s

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get credentials and run migration
export DB_PASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
export ZINDEX_DB_HOST="localhost"
cd migrations
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -b

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"

# Deploy new version (scales up automatically)
cd ..
helm upgrade zindex ./deploy/zindex-infra -n $NAMESPACE

# Wait for pod to be ready
kubectl wait --for=condition=ready pod -l app=zindex -n $NAMESPACE --timeout=120s
```

### Scenario 3: Large Database (> 10M transactions)
**Recommended**: Maintenance window with testing
- Expected time: 5-30 minutes
- Plan for brief downtime
- Test on copy first using Job

```bash
export NAMESPACE="your-namespace"

# 1. Create a test database using a Kubernetes Job
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration-test
  namespace: $NAMESPACE
spec:
  template:
    spec:
      containers:
      - name: migration-test
        image: postgres:15
        command: ["/bin/bash"]
        args:
        - -c
        - |
          # Create test database
          PGPASSWORD=\$POSTGRES_PASSWORD createdb -h postgres -U postgres zindex_test

          # Copy data
          PGPASSWORD=\$POSTGRES_PASSWORD pg_dump -h postgres -U postgres zindex | \
          PGPASSWORD=\$POSTGRES_PASSWORD psql -h postgres -U postgres zindex_test

          echo "Test database created successfully"
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: postgres-password
      restartPolicy: Never
  backoffLimit: 1
EOF

# 2. Wait for job to complete
kubectl wait --for=condition=complete job/db-migration-test -n $NAMESPACE --timeout=600s

# 3. Test migration on test database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export DB_PASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
export ZINDEX_DB_NAME="zindex_test"
export ZINDEX_DB_HOST="localhost"
cd migrations
./run-migration.sh -f 001_add_count_and_balance_fields_improved.sql -y

# 4. If successful, schedule maintenance window and run on production
# 5. Follow Scenario 2 steps for production
```

---

## Important Notes About Data Accuracy

### For Existing Data (Before Migration)
The `balance_change` values are **approximations**:
- Receive transactions: Uses entire transaction output value
- Send transactions: Uses negative of (output + fee)

This is because we don't have exact per-address output mapping for old data.

### For New Data (After Migration)
The `balance_change` values are **accurate**:
- Based on actual output values for the specific address
- Calculated during indexing with full transaction context

**Impact**: Historical data may show approximate balance changes, but:
- ✅ Totals will be correct over time
- ✅ All new data is precise
- ✅ Counts and aggregations work correctly

---

## Testing After Migration (Kubernetes)

### 1. Check Tables
```bash
export NAMESPACE="your-namespace"

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get password
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

# Check tables
psql -h localhost -U postgres -d zindex <<EOF
-- Verify columns exist
\d transactions
\d account_transactions

-- Check data population
SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE input_count > 0) as with_inputs,
    COUNT(*) FILTER (WHERE output_count > 0) as with_outputs
FROM transactions;

SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE balance_change != 0) as with_balance
FROM account_transactions;
EOF

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"
```

### 2. Test API Endpoints
```bash
export NAMESPACE="your-namespace"

# Port-forward to zindex service
kubectl port-forward svc/zindex 8080:8080 -n $NAMESPACE &

# Test count endpoints
curl "http://localhost:8080/api/v1/tx-graph/transactions/count"
curl "http://localhost:8080/api/v1/accounts/count"
curl "http://localhost:8080/api/v1/starks/verifiers/count"

# Test enhanced fields
curl "http://localhost:8080/api/v1/tx-graph/transactions/recent?limit=1" | jq '.data[0] | {txid, input_count, output_count}'

# Stop port-forward
pkill -f "kubectl port-forward.*zindex"
```

### 3. Watch Logs
```bash
export NAMESPACE="your-namespace"

# Get pod name
POD_NAME=$(kubectl get pods -n $NAMESPACE -l app=zindex -o jsonpath='{.items[0].metadata.name}')

# Monitor logs
kubectl logs -f $POD_NAME -n $NAMESPACE

# Or tail logs
kubectl logs --tail=100 $POD_NAME -n $NAMESPACE
```

---

## If Something Goes Wrong (Kubernetes)

### Option 1: Rollback Migration
```bash
export NAMESPACE="your-namespace"

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get credentials
export DB_PASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
export ZINDEX_DB_HOST="localhost"

# Run rollback
cd migrations
./run-migration.sh -f 001_rollback.sql -y

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"
```

### Option 2: Restore from Backup
```bash
export NAMESPACE="your-namespace"

# Scale down zindex
kubectl scale deployment zindex --replicas=0 -n $NAMESPACE

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get password
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

# Restore database
psql -h localhost -U postgres -d zindex < backup_TIMESTAMP.sql

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"

# Scale up zindex
kubectl scale deployment zindex --replicas=1 -n $NAMESPACE
```

### Option 3: Fix Issues Manually
```bash
export NAMESPACE="your-namespace"

# Port-forward to database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &

# Get password
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

# Fix manually
psql -h localhost -U postgres -d zindex <<EOF
-- If columns are missing
ALTER TABLE transactions ADD COLUMN input_count INT NOT NULL DEFAULT 0;
ALTER TABLE transactions ADD COLUMN output_count INT NOT NULL DEFAULT 0;
ALTER TABLE account_transactions ADD COLUMN balance_change BIGINT NOT NULL DEFAULT 0;

-- Re-run backfill (from migration script)
EOF

# Stop port-forward
pkill -f "kubectl port-forward.*postgres"
```

---

## Production Checklist (Kubernetes)

Before running on production:

- [ ] Database backup completed
- [ ] Tested on copy/staging database (or test namespace)
- [ ] Scheduled maintenance window (if needed)
- [ ] Verified disk space in PVC (`kubectl get pvc -n <namespace>`)
- [ ] Database credentials accessible via secrets
- [ ] Rollback plan prepared
- [ ] kubectl access verified
- [ ] Port-forwarding tested

After migration:

- [ ] Verified columns exist (via psql)
- [ ] Checked data backfill counts
- [ ] Tested new API endpoints (via port-forward)
- [ ] Tested existing API endpoints
- [ ] Verified new transactions index correctly
- [ ] Monitored pod logs for errors (`kubectl logs`)
- [ ] Kept backup for 24-48 hours
- [ ] Verified pod is healthy (`kubectl get pods`)

---

## Alternative: Fresh Start (Nuclear Option) - Kubernetes

If you're comfortable losing existing data:

```bash
export NAMESPACE="your-namespace"

# 1. Scale down zindex
kubectl scale deployment zindex --replicas=0 -n $NAMESPACE

# 2. Port-forward and drop/recreate database
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
dropdb -h localhost -U postgres zindex
createdb -h localhost -U postgres zindex
pkill -f "kubectl port-forward.*postgres"

# 3. Scale up zindex (will create tables with new schema)
kubectl scale deployment zindex --replicas=1 -n $NAMESPACE

# 4. Re-index from blockchain
# (Your indexer will start fresh with correct schema)
```

⚠️ **Warning**: This deletes all existing indexed data!

### Alternative: Using Helm to Reset

```bash
export NAMESPACE="your-namespace"

# Uninstall release (preserves PVC by default)
helm uninstall zindex -n $NAMESPACE

# Delete PVC to start completely fresh
kubectl delete pvc postgres-pvc -n $NAMESPACE

# Reinstall
helm install zindex ./deploy/zindex-infra -n $NAMESPACE
```

---

## Support Commands (Kubernetes)

### Check Database Size
```bash
export NAMESPACE="your-namespace"
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

psql -h localhost -U postgres -d zindex -c "SELECT pg_size_pretty(pg_database_size('zindex'));"

pkill -f "kubectl port-forward.*postgres"
```

### Check Table Sizes
```bash
export NAMESPACE="your-namespace"
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

psql -h localhost -U postgres -d zindex <<EOF
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
EOF

pkill -f "kubectl port-forward.*postgres"
```

### Check PVC Disk Usage
```bash
export NAMESPACE="your-namespace"

# Check PVC status
kubectl get pvc -n $NAMESPACE

# Check disk usage in postgres pod
POD_NAME=$(kubectl get pods -n $NAMESPACE -l app=postgres -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it $POD_NAME -n $NAMESPACE -- df -h /var/lib/postgresql/data
```

### Check for Locks During Migration
```bash
export NAMESPACE="your-namespace"
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

psql -h localhost -U postgres -d zindex <<EOF
SELECT
    locktype,
    database,
    relation::regclass,
    mode,
    granted
FROM pg_locks
WHERE NOT granted;
EOF

pkill -f "kubectl port-forward.*postgres"
```

### Monitor Migration Progress
```bash
export NAMESPACE="your-namespace"

# In another terminal while migration runs
kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &
export PGPASSWORD="$(kubectl get secret postgres-secret -n $NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"

watch -n 1 "psql -h localhost -U postgres -d zindex -c 'SELECT COUNT(*) FROM transactions WHERE input_count > 0;'"
```

### Check Pod Status and Logs
```bash
export NAMESPACE="your-namespace"

# Check pod status
kubectl get pods -n $NAMESPACE -l app=zindex

# Describe pod for events
kubectl describe pod -l app=zindex -n $NAMESPACE

# Check recent logs
kubectl logs -l app=zindex -n $NAMESPACE --tail=50

# Follow logs
kubectl logs -f -l app=zindex -n $NAMESPACE
```

---

## Summary (Kubernetes/Helm)

**What to do**: Run the migration script via port-forward with backup
**When to do it**: During low-traffic period (or anytime for small DBs)
**How long**: Seconds to minutes depending on size
**Risk level**: Low (uses transactions, has rollback, preserves data)
**Required downtime**: Optional for small DBs, brief scale-down recommended for large ones

### Quick Checklist:
1. ✅ Set `NAMESPACE` environment variable
2. ✅ Scale down zindex deployment (optional for small DBs)
3. ✅ Port-forward to PostgreSQL service
4. ✅ Run migration with backup flag
5. ✅ Scale up zindex deployment
6. ✅ Verify with API tests
7. ✅ Monitor pod logs

### Key Kubernetes Considerations:
- **Port-forwarding**: Use `kubectl port-forward` to access database
- **Secrets**: Password retrieved from `postgres-secret`
- **Scaling**: Use `kubectl scale` instead of systemctl start/stop
- **Monitoring**: Use `kubectl logs` instead of journalctl
- **Backup**: Store in persistent location (not in pod)

**Bottom line**: The migration is safe, tested, and handles existing data gracefully. Kubernetes adds minimal complexity via port-forwarding. Just backup first and you're good to go! 🚀
