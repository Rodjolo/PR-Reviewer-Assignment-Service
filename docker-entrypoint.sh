#!/bin/sh
set -e

echo "Waiting for database to be ready..."
sleep 2

echo "Running migrations..."
export MIGRATIONS_PATH="file:///root/migrations"
./migrate -up

echo "Starting server..."
exec ./server

