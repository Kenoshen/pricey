-- setup.sql
-- Run this once as a PostgreSQL superuser before starting the application.
-- Example: psql -U postgres -f setup.sql
--
-- This script is idempotent — safe to run multiple times.

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'prycebook') THEN
        CREATE ROLE prycebook LOGIN;
    END IF;
END
$$;

SELECT 'CREATE DATABASE prycebook OWNER prycebook'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'prycebook')\gexec
