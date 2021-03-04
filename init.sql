/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

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
                    order_id INT,
                    LEYE_id INT DEFAULT -1
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
    paid_for BOOL DEFAULT false,
    total_splits INT default 0,
    paid_by TEXT[] DEFAULT '{}',
    selected_by text DEFAULT '',
    split_by TEXT[] DEFAULT '{}',
    selected_by_lock bool default false
);

create table tx (
  item_id INT,
  paid_by TEXT,
  tip FLOAT default 0.0
);

