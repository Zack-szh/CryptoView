CREATE TABLE IF NOT EXISTS trades (
    time        TIMESTAMPTZ     NOT NULL,
    ticker      TEXT            NOT NULL, 
    price       DECIMAL(20, 8)  NOT NULL,
    quantity    DECIMAL(20, 8)  NOT NULL,
    is_maker    BOOLEAN         NOT NULL,
    trade_id    BIGINT          NOT NULL,
    PRIMARY KEY(time, ticker, trade_id)
)

