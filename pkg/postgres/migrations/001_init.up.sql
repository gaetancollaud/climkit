CREATE TABLE api_token
(
    api_token_id VARCHAR PRIMARY KEY NOT NULL,
    expire_at    TIMESTAMP           NOT NULL,
    content      VARCHAR             NOT NULL
);

CREATE TABLE installations
(
    installation_id VARCHAR PRIMARY KEY NOT NULL,
    site_ref        VARCHAR             NOT NULL,
    name            VARCHAR             NOT NULL,
    timezone        VARCHAR             NOT NULL,
    latitude        DOUBLE PRECISION,
    longitude       DOUBLE PRECISION
);

CREATE TABLE meters
(
    meter_id        VARCHAR PRIMARY KEY NOT NULL,
    installation_id VARCHAR             NOT NULL,
    meter_type      VARCHAR             NOT NULL,
    prim_ad         INTEGER             NOT NULL,
    virtual         BOOLEAN             NOT NULL,
    CONSTRAINT meters_installation_id
        FOREIGN KEY (installation_id)
            REFERENCES installations (installation_id)
);

CREATE TABLE installation_values
(
    installation_id VARCHAR          NOT NULL,
    date_time       TIMESTAMP        NOT NULL,
    prod_total      DOUBLE PRECISION NOT NULL,
    self            DOUBLE PRECISION NOT NULL,
    to_ext          DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (installation_id, date_time),
    CONSTRAINT installation_values_installation_id
        FOREIGN KEY (installation_id)
            REFERENCES installations (installation_id)
);

CREATE TABLE meter_values
(
    meter_id  VARCHAR          NOT NULL,
    date_time TIMESTAMP        NOT NULL,
    total     DOUBLE PRECISION NOT NULL,
    self      DOUBLE PRECISION NOT NULL,
    ext       DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (meter_id, date_time),
    CONSTRAINT meter_values_meter_id
        FOREIGN KEY (meter_id)
            REFERENCES meters (meter_id)
);