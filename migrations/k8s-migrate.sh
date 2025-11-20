#!/bin/bash

# Kubernetes Migration Helper Script
# This script automates the migration process for zindex running in Kubernetes

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 command not found. Please install it first."
        exit 1
    fi
}

# Function to cleanup on exit
cleanup() {
    print_info "Cleaning up..."
    pkill -f "kubectl port-forward.*postgres" 2>/dev/null || true

    if [ "$SCALED_DOWN" = true ] && [ "$SKIP_SCALE_UP" != true ]; then
        print_info "Scaling zindex back up..."
        kubectl scale deployment zindex --replicas=1 -n "$NAMESPACE" 2>/dev/null || true
    fi
}

trap cleanup EXIT

# Main script
echo "=========================================="
echo "Zindex Kubernetes Migration Helper"
echo "=========================================="
echo ""

# Check for required commands
check_command kubectl
check_command psql
check_command pg_dump

# Parse command line arguments
NAMESPACE=""
MIGRATION_FILE="001_add_count_and_balance_fields_improved.sql"
CREATE_BACKUP=true
SCALE_DOWN=false
SKIP_CONFIRMATION=false
SCALED_DOWN=false
SKIP_SCALE_UP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -f|--file)
            MIGRATION_FILE="$2"
            shift 2
            ;;
        --no-backup)
            CREATE_BACKUP=false
            shift
            ;;
        --scale-down)
            SCALE_DOWN=true
            shift
            ;;
        -y|--yes)
            SKIP_CONFIRMATION=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -n, --namespace NS    Kubernetes namespace (required)"
            echo "  -f, --file FILE       Migration file (default: 001_add_count_and_balance_fields_improved.sql)"
            echo "  --no-backup           Skip database backup"
            echo "  --scale-down          Scale down zindex deployment during migration"
            echo "  -y, --yes             Skip confirmation prompt"
            echo "  -h, --help            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 -n zindex-prod"
            echo "  $0 -n zindex-prod --scale-down"
            echo "  $0 -n zindex-prod -f 001_rollback.sql --no-backup -y"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Check if namespace is provided
if [ -z "$NAMESPACE" ]; then
    print_error "Namespace not specified. Use -n or --namespace option."
    echo "Use -h or --help for usage information"
    exit 1
fi

# Check if migration file exists
if [ ! -f "$MIGRATION_FILE" ]; then
    print_error "Migration file not found: $MIGRATION_FILE"
    exit 1
fi

print_info "Configuration:"
echo "  Namespace: $NAMESPACE"
echo "  Migration file: $MIGRATION_FILE"
echo "  Backup: $([ "$CREATE_BACKUP" = true ] && echo "Yes" || echo "No")"
echo "  Scale down: $([ "$SCALE_DOWN" = true ] && echo "Yes" || echo "No")"
echo ""

# Confirmation prompt
if [ "$SKIP_CONFIRMATION" = false ]; then
    read -p "Do you want to proceed? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        print_info "Migration cancelled."
        exit 0
    fi
fi

echo ""
print_info "Starting migration process..."
echo ""

# Step 1: Scale down if requested
if [ "$SCALE_DOWN" = true ]; then
    print_info "Scaling down zindex deployment..."
    kubectl scale deployment zindex --replicas=0 -n "$NAMESPACE"
    SCALED_DOWN=true

    print_info "Waiting for pods to terminate..."
    kubectl wait --for=delete pod -l app=zindex -n "$NAMESPACE" --timeout=60s || true
    echo ""
fi

# Step 2: Port-forward to PostgreSQL
print_info "Setting up port-forward to PostgreSQL..."
kubectl port-forward svc/postgres 5432:5432 -n "$NAMESPACE" &>/dev/null &
PORT_FORWARD_PID=$!

# Wait for port-forward to be ready
sleep 3

if ! kill -0 $PORT_FORWARD_PID 2>/dev/null; then
    print_error "Failed to establish port-forward to PostgreSQL"
    exit 1
fi

print_success "Port-forward established"
echo ""

# Step 3: Get database credentials
print_info "Retrieving database credentials from secrets..."
export DB_PASSWORD=$(kubectl get secret postgres-secret -n "$NAMESPACE" -o jsonpath='{.data.postgres-password}' 2>/dev/null | base64 -d)

if [ -z "$DB_PASSWORD" ]; then
    print_error "Failed to retrieve database password from secret 'postgres-secret'"
    exit 1
fi

export ZINDEX_DB_HOST="localhost"
export ZINDEX_DB_PORT="5432"
export ZINDEX_DB_NAME="zindex"
export ZINDEX_DB_USER="postgres"

print_success "Credentials retrieved"
echo ""

# Step 4: Create backup if requested
if [ "$CREATE_BACKUP" = true ]; then
    print_info "Creating database backup..."
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"

    PGPASSWORD="$DB_PASSWORD" pg_dump -h localhost -p 5432 -U postgres zindex > "$BACKUP_FILE" 2>/dev/null

    if [ $? -eq 0 ]; then
        print_success "Backup created: $BACKUP_FILE"
        echo ""
    else
        print_error "Backup failed!"
        exit 1
    fi
fi

# Step 5: Run the migration
print_info "Running migration: $MIGRATION_FILE"
echo ""

./run-migration.sh -f "$MIGRATION_FILE" -y

if [ $? -ne 0 ]; then
    print_error "Migration failed!"
    exit 1
fi

echo ""
print_success "Migration completed successfully!"
echo ""

# Step 6: Cleanup and scale up
print_info "Stopping port-forward..."
pkill -f "kubectl port-forward.*postgres" || true
echo ""

if [ "$SCALED_DOWN" = true ]; then
    print_info "Scaling zindex back up..."
    kubectl scale deployment zindex --replicas=1 -n "$NAMESPACE"

    print_info "Waiting for pod to be ready..."
    kubectl wait --for=condition=ready pod -l app=zindex -n "$NAMESPACE" --timeout=120s

    SKIP_SCALE_UP=true  # Prevent cleanup from scaling up again

    print_success "Zindex deployment is ready"
    echo ""
fi

# Step 7: Verification
print_info "Verifying deployment..."
POD_NAME=$(kubectl get pods -n "$NAMESPACE" -l app=zindex -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)

if [ -n "$POD_NAME" ]; then
    print_success "Zindex pod is running: $POD_NAME"

    print_info "Recent logs:"
    kubectl logs "$POD_NAME" -n "$NAMESPACE" --tail=10 2>/dev/null || print_warning "Could not retrieve logs"
else
    print_warning "No zindex pods found"
fi

echo ""
echo "=========================================="
print_success "Migration process complete!"
echo "=========================================="
echo ""

if [ "$CREATE_BACKUP" = true ]; then
    print_info "Backup file: $BACKUP_FILE"
    print_info "To restore: kubectl port-forward svc/postgres 5432:5432 -n $NAMESPACE &"
    print_info "            psql -h localhost -U postgres -d zindex < $BACKUP_FILE"
    echo ""
fi

print_info "Next steps:"
echo "  1. Test the API endpoints (see test commands in curl-test-commands.txt)"
echo "  2. Monitor logs: kubectl logs -f -l app=zindex -n $NAMESPACE"
echo "  3. Keep the backup for 24-48 hours"
echo ""
print_success "Done! 🚀"
