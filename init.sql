create table USERS (
                       phone TEXT,
                       fname TEXT,
                       lname TEXT,
                       auth_token TEXT,
                       auth_token_exp TIME,
                       primary_card TEXT,
                       current_order INT,
                       current_sms_verification_token TEXT,
                       past_orders INT[] DEFAULT '{}'
);

/* @TODO: do more secure */
create table PAYINFO (
                         phone TEXT,
                         fname TEXT,
                         lname TEXT,
                         num TEXT,
                         cvv INT,
                         exp TEXT
);


create table TOKENS (
                    token_code TEXT,
                    rest_name TEXT,
                    rest_id INT,
                    table_id INT,
                    order_id INT
);

create table ORDERS (
    order_id INT,
    rest_name TEXT
);


create table ORDERITEMS (
    order_id INT,
    item_name TEXT,
    item_type TEXT,
    item_cost FLOAT,
    item_id SERIAL,
    paid_for BOOL,
    total_splits INT,
    paid_by TEXT[] DEFAULT '{}',
    selected_by text DEFAULT '',
    split_by TEXT[] DEFAULT '{}',
    selected_by_lock bool default false
);


