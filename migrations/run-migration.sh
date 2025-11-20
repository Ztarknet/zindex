#!/bin/bash

# Migration runner script
# This script helps you run database migrations for zindex

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default database connection settings
# Override these with environment variables if needed
DB_HOST="${ZINDEX_DB_HOST:-localhost}"
DB_PORT="${ZINDEX_DB_PORT:-5432}"
DB_NAME="${ZINDEX_DB_NAME:-zindex}"
DB_USER="${ZINDEX_DB_USER:-postgres}"

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

# Function to check if psql is available
check_psql() {
    if ! command -v psql &> /dev/null; then
        print_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi
}

# Function to run a migration file
run_migration() {
    local migration_file=$1

    if [ ! -f "$migration_file" ]; then
        print_error "Migration file not found: $migration_file"
        exit 1
    fi

    print_info "Running migration: $migration_file"
    print_info "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
    echo ""

    # Run the migration
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"

    if [ $? -eq 0 ]; then
        print_success "Migration completed successfully!"
    else
        print_error "Migration failed!"
        exit 1
    fi
}

# Function to create a backup
create_backup() {
    print_info "Creating database backup..."

    local backup_file="backup_$(date +%Y%m%d_%H%M%S).sql"

    PGPASSWORD="$DB_PASSWORD" pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" > "$backup_file"

    if [ $? -eq 0 ]; then
        print_success "Backup created: $backup_file"
        echo "$backup_file"
    else
        print_error "Backup failed!"
        exit 1
    fi
}

# Main script
echo "=========================================="
echo "Zindex Database Migration Runner"
echo "=========================================="
echo ""

# Check for psql
check_psql

# Parse command line arguments
MIGRATION_FILE=""
CREATE_BACKUP=false
SKIP_CONFIRMATION=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--file)
            MIGRATION_FILE="$2"
            shift 2
            ;;
        -b|--backup)
            CREATE_BACKUP=true
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
            echo "  -f, --file FILE       Migration file to run (required)"
            echo "  -b, --backup          Create backup before migration"
            echo "  -y, --yes             Skip confirmation prompt"
            echo "  -h, --help            Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  ZINDEX_DB_HOST        Database host (default: localhost)"
            echo "  ZINDEX_DB_PORT        Database port (default: 5432)"
            echo "  ZINDEX_DB_NAME        Database name (default: zindex)"
            echo "  ZINDEX_DB_USER        Database user (default: postgres)"
            echo "  DB_PASSWORD           Database password (required)"
            echo ""
            echo "Examples:"
            echo "  $0 -f 001_add_count_and_balance_fields_improved.sql"
            echo "  $0 -f 001_add_count_and_balance_fields_improved.sql -b"
            echo "  $0 -f 001_rollback.sql -b -y"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Check if migration file is provided
if [ -z "$MIGRATION_FILE" ]; then
    print_error "No migration file specified. Use -f or --file option."
    echo "Use -h or --help for usage information"
    exit 1
fi

# Check if password is provided
if [ -z "$DB_PASSWORD" ]; then
    print_warning "DB_PASSWORD environment variable not set."
    read -sp "Enter database password: " DB_PASSWORD
    echo ""
    export DB_PASSWORD
fi

# Display migration info
print_info "Migration file: $MIGRATION_FILE"
print_info "Database: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
print_info "Backup: $([ "$CREATE_BACKUP" = true ] && echo "Yes" || echo "No")"
echo ""

# Confirmation prompt
if [ "$SKIP_CONFIRMATION" = false ]; then
    read -p "Do you want to proceed? (yes/no): " confirm
    if [ "$confirm" != "yes" ]; then
        print_info "Migration cancelled."
        exit 0
    fi
fi

# Create backup if requested
if [ "$CREATE_BACKUP" = true ]; then
    BACKUP_FILE=$(create_backup)
    echo ""
fi

# Run the migration
run_migration "$MIGRATION_FILE"

echo ""
print_success "All done!"

if [ "$CREATE_BACKUP" = true ]; then
    echo ""
    print_info "Backup file: $BACKUP_FILE"
    print_info "To restore: psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME < $BACKUP_FILE"
fi

echo ""
echo "=========================================="
