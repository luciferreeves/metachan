#!/bin/bash

DEFAULT_USER=$(whoami)

read -p "Enter database name [metachan]: " DB_NAME
DB_NAME=${DB_NAME:-metachan}

read -p "Enter database owner username [$DEFAULT_USER]: " DB_OWNER
DB_OWNER=${DB_OWNER:-$DEFAULT_USER}

read -s -p "Enter password for $DB_OWNER (press Enter for no password): " DB_PASSWORD
echo

if [ -z "$DB_PASSWORD" ]; then
    PSQL_CMD="psql -d postgres"
    CONNECTION_INFO="Connecting as current user without password"
else
    export PGPASSWORD="$DB_PASSWORD"
    PSQL_CMD="psql -U $DB_OWNER -d postgres"
    CONNECTION_INFO="Connecting as $DB_OWNER with password"
fi

DB_EXISTS=$($PSQL_CMD -t -c "SELECT 1 FROM pg_database WHERE datname='$DB_NAME';" 2>/dev/null | xargs)

if [ "$DB_EXISTS" = "1" ]; then
    ACTION="reset"
    OPERATION="Database reset"
else
    ACTION="create"
    OPERATION="Database creation"
fi

echo
echo "=== $OPERATION Configuration ==="
echo "Database name: $DB_NAME"
echo "Owner: $DB_OWNER"
echo "Connection: $CONNECTION_INFO"
if [ "$ACTION" = "reset" ]; then
    echo "Status: Database exists and will be reset"
else
    echo "Status: Database does not exist and will be created"
fi
echo

read -p "Proceed with database $ACTION? [Y/n]: " CONFIRM
if [[ $CONFIRM =~ ^[Nn]$ ]]; then
    echo "Operation cancelled."
    exit 0
fi

if [ "$ACTION" = "reset" ]; then
    echo "Resetting database '$DB_NAME'..."
else
    echo "Creating database '$DB_NAME'..."
fi

$PSQL_CMD -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null
if $PSQL_CMD -c "CREATE DATABASE $DB_NAME OWNER $DB_OWNER;" 2>/dev/null; then
    if [ "$ACTION" = "reset" ]; then
        echo "✅ Database '$DB_NAME' successfully reset with owner '$DB_OWNER'"
    else
        echo "✅ Database '$DB_NAME' successfully created with owner '$DB_OWNER'"
    fi
    echo
    echo "Database connection DSN:"
    if [ -z "$DB_PASSWORD" ]; then
        DSN="postgresql://$DB_OWNER@localhost:5432/$DB_NAME?sslmode=disable"
    else
        DSN="postgresql://$DB_OWNER:$DB_PASSWORD@localhost:5432/$DB_NAME?sslmode=disable"
    fi
    echo "$DSN"
else
    echo "❌ Error: Failed to $ACTION database. Please check:"
    echo "  - PostgreSQL is running"
    echo "  - User '$DB_OWNER' exists and has necessary privileges"
    echo "  - No active connections to database '$DB_NAME'"
    exit 1
fi

unset PGPASSWORD