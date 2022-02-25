SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS account
(
    id        UUID UNIQUE PRIMARY KEY,
    balance   BIGINT CHECK ( balance >= 0 ),
    createdAt TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS index_name
    ON account (balance);
CREATE INDEX IF NOT EXISTS index_id_balance
    ON account (id, balance);

CREATE INDEX IF NOT EXISTS index_date_ft
    ON financial_transaction (datetimestamp);

CREATE INDEX IF NOT EXISTS index_amount_ft
    ON financial_transaction (amount);

CREATE TABLE IF NOT EXISTS financial_transaction
(
    id            UUID UNIQUE PRIMARY KEY,
    datetimestamp TIMESTAMP NOT NULL           DEFAULT NOW(),
    description   TEXT      NOT NULL,
    from_id       UUID REFERENCES account (id) DEFAULT NULL,
    to_id         UUID REFERENCES account (id) DEFAULT NULL,
    amount        BIGINT
);
