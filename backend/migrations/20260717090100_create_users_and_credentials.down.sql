DROP TABLE IF EXISTS auth_action_tokens;

DROP TRIGGER IF EXISTS trg_user_passwords_touch_updated_at ON user_passwords;

DROP TABLE IF EXISTS user_passwords;

DROP TRIGGER IF EXISTS trg_users_touch_updated_at ON users;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS auth_action_token_purpose;

DROP TYPE IF EXISTS user_status;