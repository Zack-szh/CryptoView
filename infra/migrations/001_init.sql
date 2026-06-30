CREATE TABLE IF NOT EXISTS trades (
    time        TIMESTAMPTZ     NOT NULL,
    ticker      TEXT            NOT NULL, 
    price       DECIMAL(8, 20)  NOT NULL,
    quantity    DECIMAL(8, 20)  NOT NULL,
    side        TEXT            NOT NULL,
    trade_id    BIGINT          NOT NULL,
    PRIMARY KEY(time, ticker, trade_id)
)

