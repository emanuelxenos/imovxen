DROP TRIGGER IF EXISTS trg_sync_user_listings ON properties;
DROP FUNCTION IF EXISTS fn_sync_user_listings_count();

DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS property_images;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS "pgcrypto";
