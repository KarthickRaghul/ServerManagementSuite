#!/bin/sh

echo "⏳ Waiting for DB to be ready..."

until nc -z -v -w5 sms-db 5432; do
  echo "🛢️  Waiting for postgres container..."
  sleep 2
done

echo "⏳ Running DB init..."
./dbinit

echo "🚀 Starting server..."
exec ./server
