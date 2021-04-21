/*
 * This file is subject to the additional terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 * Copyright 2020-2021 Dominic "vopi181" Pace
 */

create table USERS (
                       phone TEXT PRIMARY KEY,
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
                    token_code TEXT PRIMARY KEY,
                    rest_id INT,
                    table_id INT,
                    order_id INT
);

create table RESTAURANTS (
    rest_id SERIAL PRIMARY KEY,
    rest_name TEXT NOT NULL,
    menu_url TEXT,
    LEYE_id INT,
    zip TEXT NOT NULL
);

create table ORDERS (
    order_id INT PRIMARY KEY,
    rest_id INT
);


create table ORDERITEMS (
    order_id INT,
    item_name TEXT,
    item_type TEXT,
    item_cost FLOAT,
    item_id SERIAL PRIMARY KEY,
    paid_for BOOL DEFAULT false,
    total_splits INT default 0,
    paid_by TEXT[] DEFAULT '{}',
    selected_by text DEFAULT '',
    split_by TEXT[] DEFAULT '{}',
    selected_by_lock bool default false
);

create table tx (
  order_id INT NOT NULL,
  paid_by TEXT NOT NULL,
  tx_id SERIAL PRIMARY KEY,
  tip FLOAT default 0.0,
  LEYE_pin INT DEFAULT 0,
  LEYE_bal FLOAT DEFAULT 0.0,
  device_info TEXT DEFAULT '',
  geo_id TEXT DEFAULT '',
  time timestamptz NOT NULL,
  user_authed BOOL NOT NULL DEFAULT false
);

