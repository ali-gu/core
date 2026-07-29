-- +goose Up

CREATE TABLE entity_base (
    entity_state TEXT NOT NULL
);

CREATE TABLE practices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    email TEXT,
    zip_code TEXT,
    website TEXT
) INHERITS (entity_base);

CREATE TABLE roles (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    disabled_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE locations (
    id TEXT PRIMARY KEY,
    address TEXT NOT NULL,
    practice_id TEXT NOT NULL REFERENCES practices(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE phone_numbers (
    id TEXT PRIMARY KEY,
    number TEXT NOT NULL,
    phone_number_ref TEXT,
    phone_number_reservation_ref TEXT,
    practice_id TEXT NOT NULL REFERENCES practices(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE agents (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    phone_number_id TEXT REFERENCES phone_numbers(id),
    location_id TEXT REFERENCES locations(id),
    agent_ref TEXT,
    practice_id TEXT NOT NULL REFERENCES practices(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE ehrs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    subdomain TEXT,
    location_ref TEXT,
    location_id TEXT NOT NULL REFERENCES locations(id),
    created_at TIMESTAMPTZ NOT NULL,
    onboarding_url TEXT NOT NULL,
    onboarding_id TEXT NOT NULL
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    user_ref TEXT NOT NULL,
    role_id TEXT NOT NULL REFERENCES roles(id),
    practice_id TEXT NOT NULL REFERENCES practices(id),
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE tools (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    tool_ref TEXT NOT NULL DEFAULT '',
    config JSONB NOT NULL DEFAULT '{}'::jsonb,
    kind TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ
) INHERITS (entity_base);

CREATE TABLE conversations (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES agents(id),
    phone_number_id TEXT NOT NULL REFERENCES phone_numbers(id),
    location_id TEXT NOT NULL REFERENCES locations(id),
    practice_id TEXT NOT NULL REFERENCES practices(id),
    conversation_ref TEXT NOT NULL,
    duration BIGINT NOT NULL DEFAULT 0,
    outcome TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE nexhealth_auth_tokens (
    id TEXT PRIMARY KEY,
    token TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL DEFAULT 'epoch',
    updated_at TIMESTAMPTZ
);

CREATE TABLE http_requests (
    id            TEXT PRIMARY KEY,
    practice_id   TEXT REFERENCES practices(id),
    user_id       TEXT REFERENCES users(id),
    method        TEXT NOT NULL,
    path          TEXT NOT NULL,
    query_params  TEXT,
    headers       JSONB,
    ip_address    TEXT NOT NULL,
    request_body  TEXT,
    response_body TEXT,
    status_code   INTEGER NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL
);

INSERT INTO roles (entity_state, id, type, created_at) VALUES
    ('ACTIVE', '3GFBfoEXnvJRJqEUHbXoqIoJ0GQ', 'SUPER_ADMIN', now()),
    ('ACTIVE', '3GFBftnFnROXjlfSex0EaOD2QaK', 'ADMIN', now()),
    ('ACTIVE', '3GFBfsvt8iIsFC8zi7nfIhjUoA5', 'READER', now());

INSERT INTO nexhealth_auth_tokens (id, expires_at) VALUES ('singleton', 'epoch');

-- +goose Down

DROP TABLE http_requests;
DROP TABLE nexhealth_auth_tokens;
DROP TABLE conversations;
DROP TABLE tools;
DROP TABLE users;
DROP TABLE ehrs;
DROP TABLE agents;
DROP TABLE phone_numbers;
DROP TABLE locations;
DROP TABLE roles;
DROP TABLE practices;
DROP TABLE entity_base;
