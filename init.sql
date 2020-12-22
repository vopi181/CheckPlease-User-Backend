create table USERS (
                       phone TEXT,
                       fname TEXT,
                       lname TEXT,
                       auth_token TEXT,
                       auth_token_exp TIME,
                       primary_card TEXT,
                       current_order INT,
                       current_sms_verification_token TEXT
);

/* Temp. do more secure */
create table PAYINFO (
                         num TEXT,
                         cvv INT,
                         exp TEXT
);


