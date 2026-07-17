CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE FUNCTION touch_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

COMMENT ON FUNCTION touch_updated_at () IS 'Sets updated_at to the current transaction timestamp before a row update.';