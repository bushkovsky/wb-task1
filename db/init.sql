CREATE TABLE orders (
                        order_uid VARCHAR PRIMARY KEY,
                        track_number VARCHAR,
                        entry VARCHAR,
                        locale VARCHAR,
                        customer_id VARCHAR,
                        delivery_service VARCHAR,
                        shardkey VARCHAR,
                        sm_id INTEGER,
                        date_created TIMESTAMP,
                        oof_shard VARCHAR
);

ALTER TABLE orders
    ADD COLUMN frequency INTEGER;

CREATE TABLE delivery (
                          order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid),
                          name VARCHAR,
                          phone VARCHAR,
                          zip VARCHAR,
                          city VARCHAR,
                          address VARCHAR,
                          region VARCHAR,
                          email VARCHAR
);

CREATE TABLE payment (
                         order_uid VARCHAR PRIMARY KEY REFERENCES orders(order_uid),
                         transaction VARCHAR,
                         currency VARCHAR,
                         provider VARCHAR,
                         amount INTEGER,
                         payment_dt BIGINT,
                         bank VARCHAR,
                         delivery_cost INTEGER,
                         goods_total INTEGER,
                         custom_fee INTEGER
);

CREATE TABLE items (
                       chrt_id BIGINT,
                       order_uid VARCHAR REFERENCES orders(order_uid),
                       track_number VARCHAR,
                       price INTEGER,
                       rid VARCHAR,
                       name VARCHAR,
                       sale INTEGER,
                       size VARCHAR,
                       total_price INTEGER,
                       nm_id BIGINT,
                       brand VARCHAR,
                       status INTEGER
);


ALTER TABLE items ADD CONSTRAINT items_unique UNIQUE (chrt_id, order_uid);