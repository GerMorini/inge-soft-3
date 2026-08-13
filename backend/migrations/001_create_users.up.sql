CREATE TABLE IF NOT EXISTS users (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    first_name varchar(100) NOT NULL,
    last_name varchar(100) NOT NULL,
    phone varchar(16) NOT NULL,
    street varchar(120) NOT NULL,
    street_number varchar(20) NOT NULL,
    apartment varchar(20),
    city varchar(100) NOT NULL,
    province varchar(100) NOT NULL,
    username varchar(30) NOT NULL,
    email varchar(254) NOT NULL,
    password_hash text NOT NULL,

    CONSTRAINT users_username_key UNIQUE (username),
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT users_first_name_not_blank CHECK (btrim(first_name) <> ''),
    CONSTRAINT users_last_name_not_blank CHECK (btrim(last_name) <> ''),
    CONSTRAINT users_street_not_blank CHECK (btrim(street) <> ''),
    CONSTRAINT users_street_number_valid CHECK (
        btrim(street_number) <> '' AND char_length(street_number) <= 20
    ),
    CONSTRAINT users_city_not_blank CHECK (btrim(city) <> ''),
    CONSTRAINT users_province_not_blank CHECK (btrim(province) <> ''),
    CONSTRAINT users_username_canonical CHECK (
        username = lower(username) AND username ~ '^[a-z0-9._-]{3,30}$'
    ),
    CONSTRAINT users_email_canonical CHECK (email = lower(email)),
    CONSTRAINT users_phone_e164 CHECK (phone ~ '^\+[1-9][0-9]{1,14}$'),
    CONSTRAINT users_password_argon2id CHECK (password_hash LIKE '$argon2id$%')
);
