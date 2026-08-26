CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS users_public_id_unique
    ON users (public_id);
