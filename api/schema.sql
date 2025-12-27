CREATE DATABASE "notify-stock-test";

CREATE TABLE IF NOT EXISTS
    stocks (
        symbol TEXT,
        TIMESTAMP TIMESTAMP,
        open DECIMAL NOT NULL,
        CLOSE DECIMAL NOT NULL,
        high DECIMAL NOT NULL,
        low DECIMAL NOT NULL,
        PRIMARY KEY (symbol, TIMESTAMP)
    );

CREATE TABLE IF NOT EXISTS
    symbols (
        symbol TEXT PRIMARY KEY,
        short_name TEXT NOT NULL,
        long_name TEXT NOT NULL,
        market_price DECIMAL NOT NULL,
        previous_close DECIMAL NOT NULL,
        volume INTEGER,
        market_cap INTEGER
    );

ALTER TABLE symbols
ADD COLUMN currency TEXT;

CREATE TABLE IF NOT EXISTS
    sessions (
        id TEXT PRIMARY KEY,
        state TEXT NOT NULL,
        is_active BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
        expires_at TIMESTAMP NOT NULL
    );

ALTER TABLE sessions ADD COLUMN member_id UUID;

CREATE TABLE IF NOT EXISTS
    members (id UUID PRIMARY KEY);

CREATE TABLE IF NOT EXISTS
    google_members (
        id TEXT PRIMARY KEY,
        email TEXT NOT NULL,
        verified_email BOOLEAN NOT NULL,
        NAME TEXT NOT NULL,
        given_name TEXT NOT NULL,
        family_name TEXT NOT NULL,
        picture TEXT NOT NULL,
        member_id UUID NOT NULL,
        FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS
    notifications (id UUID PRIMARY KEY, member_id UUID, HOUR TIME, FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE);

CREATE TABLE IF NOT EXISTS
    notification_targets (
        id UUID PRIMARY KEY,
        symbol TEXT,
        notification_id UUID,
        FOREIGN KEY (notification_id) REFERENCES notifications (id) ON DELETE CASCADE,
        FOREIGN KEY (symbol) REFERENCES symbols (symbol) ON DELETE CASCADE
    );
