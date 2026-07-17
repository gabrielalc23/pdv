CREATE TYPE user_status AS ENUM (
    'ACTIVE',
    'SUSPENDED',
    'DISABLED'
);

CREATE TYPE auth_action_token_purpose AS ENUM (
    'EMAIL_VERIFICATION',
    'PASSWORD_RESET'
);

CREATE TABLE users (
    id UUID DEFAULT uuidv7 () NOT NULL,
    email VARCHAR(320) NOT NULL,
    email_normalized VARCHAR(320) NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    status user_status DEFAULT 'ACTIVE' NOT NULL,
    email_verified_at TIMESTAMPTZ,
    password_version BIGINT DEFAULT 1 NOT NULL,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_normalized_unique UNIQUE (email_normalized),
    CONSTRAINT users_email_not_blank CHECK (BTRIM (email) <> ''),
    CONSTRAINT users_email_normalized_not_blank CHECK (
        BTRIM (email_normalized) <> ''
    ),
    CONSTRAINT users_email_normalized_canonical CHECK (
        email_normalized = LOWER(BTRIM (email_normalized))
    ),
    CONSTRAINT users_display_name_not_blank CHECK (BTRIM (display_name) <> ''),
    CONSTRAINT users_password_version_positive CHECK (password_version > 0),
    CONSTRAINT users_email_verified_at_check CHECK (
        email_verified_at IS NULL
        OR email_verified_at >= created_at
    ),
    CONSTRAINT users_last_login_at_check CHECK (
        last_login_at IS NULL
        OR last_login_at >= created_at
    ),
    CONSTRAINT users_updated_at_check CHECK (updated_at >= created_at)
);

CREATE INDEX idx_users_status_created_at ON users (status, created_at);

CREATE INDEX idx_users_email_verified_at ON users (email_verified_at)
WHERE
    email_verified_at IS NOT NULL;

CREATE TRIGGER trg_users_touch_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE users IS 'Stores global user identities shared across organizations.';

COMMENT ON COLUMN users.id IS 'Internal UUIDv7 identifier.';

COMMENT ON COLUMN users.email IS 'User email preserving the trimmed display casing supplied by the application.';

COMMENT ON COLUMN users.email_normalized IS 'Trimmed lowercase email used for identity lookup and uniqueness.';

COMMENT ON COLUMN users.display_name IS 'Name displayed in user-facing interfaces.';

COMMENT ON COLUMN users.status IS 'Lifecycle status of the global identity.';

COMMENT ON COLUMN users.email_verified_at IS 'Timestamp at which ownership of the email was verified.';

COMMENT ON COLUMN users.password_version IS 'Monotonic version used to invalidate access after credential changes.';

COMMENT ON COLUMN users.last_login_at IS 'Timestamp of the latest successful password login.';

CREATE TABLE user_passwords (
    user_id UUID NOT NULL,
    password_hash TEXT NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT user_passwords_pkey PRIMARY KEY (user_id),
    CONSTRAINT user_passwords_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT user_passwords_password_hash_argon2id_phc CHECK (
        password_hash ~ '^\$argon2id\$v=19\$m=[1-9][0-9]*,t=[1-9][0-9]*,p=[1-9][0-9]*\$[A-Za-z0-9+/]+\$[A-Za-z0-9+/]+$'
    ),
    CONSTRAINT user_passwords_changed_at_check CHECK (changed_at >= created_at),
    CONSTRAINT user_passwords_updated_at_check CHECK (updated_at >= created_at)
);

CREATE TRIGGER trg_user_passwords_touch_updated_at
BEFORE UPDATE ON user_passwords
FOR EACH ROW
EXECUTE FUNCTION touch_updated_at();

COMMENT ON
TABLE user_passwords IS 'Stores the current Argon2id PHC password hash for each user.';

COMMENT ON COLUMN user_passwords.user_id IS 'User that owns this password credential.';

COMMENT ON COLUMN user_passwords.password_hash IS 'Complete Argon2id PHC string; plaintext passwords are never stored.';

COMMENT ON COLUMN user_passwords.changed_at IS 'Timestamp of the latest password change.';

CREATE TABLE auth_action_tokens (
    id UUID DEFAULT uuidv7 () NOT NULL,
    user_id UUID NOT NULL,
    purpose auth_action_token_purpose NOT NULL,
    secret_hash BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    requested_ip INET,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT auth_action_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT auth_action_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT auth_action_tokens_secret_hash_length CHECK (
        OCTET_LENGTH(secret_hash) = 32
    ),
    CONSTRAINT auth_action_tokens_expires_at_check CHECK (expires_at > created_at),
    CONSTRAINT auth_action_tokens_consumed_at_check CHECK (
        consumed_at IS NULL
        OR consumed_at >= created_at
    )
);

CREATE INDEX idx_auth_action_tokens_user_purpose_created_at ON auth_action_tokens (
    user_id,
    purpose,
    created_at DESC
);

CREATE INDEX idx_auth_action_tokens_unconsumed ON auth_action_tokens (user_id, purpose, expires_at)
WHERE
    consumed_at IS NULL;

COMMENT ON
TABLE auth_action_tokens IS 'Stores one-time email verification and password reset token selectors and hashes.';

COMMENT ON COLUMN auth_action_tokens.id IS 'UUIDv7 selector included in the external token representation.';

COMMENT ON COLUMN auth_action_tokens.user_id IS 'User for whom the action token was issued.';

COMMENT ON COLUMN auth_action_tokens.purpose IS 'Security action authorized by the token.';

COMMENT ON COLUMN auth_action_tokens.secret_hash IS 'Thirty-two-byte hash of the opaque secret; the raw token is never stored.';

COMMENT ON COLUMN auth_action_tokens.expires_at IS 'Timestamp after which the token cannot be consumed.';

COMMENT ON COLUMN auth_action_tokens.consumed_at IS 'Timestamp of the successful one-time consumption.';

COMMENT ON COLUMN auth_action_tokens.requested_ip IS 'Client IP recorded when the token was requested.';