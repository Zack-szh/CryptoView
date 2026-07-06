CREATE TABLE IF NOT EXISTS trades (
    time        TIMESTAMPTZ     NOT NULL,
    symbol      TEXT            NOT NULL, 
    price       DECIMAL(20, 8)  NOT NULL,
    quantity    DECIMAL(20, 8)  NOT NULL,
    is_maker    BOOLEAN         NOT NULL,
    trade_id    BIGINT          NOT NULL,
    PRIMARY KEY(time, symbol, trade_id)
);

CREATE TABLE IF NOT EXISTS tickers (
    time                TIMESTAMPTZ     NOT NULL,
    symbol              TEXT            NOT NULL, 
    last_price          DECIMAL(20, 8)  NOT NULL,
    open_price          DECIMAL(20, 8)  NOT NULL,
    high                DECIMAL(20, 8)  NOT NULL,
    low                 DECIMAL(20, 8)  NOT NULL,
    volume              DECIMAL(20, 8)  NOT NULL,
    quote_volume        DECIMAL(20, 8)  NOT NULL,
    weighted_avg_price  DECIMAL(20, 8)  NOT NULL,
    trade_count         BIGINT          NOT NULL,
    PRIMARY KEY(time, symbol)
)    