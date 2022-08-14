CREATE TABLE t_installations
(
    installation_id VARCHAR PRIMARY KEY NOT NULL,
    site_ref        VARCHAR             NOT NULL,
    name            VARCHAR             NOT NULL,
    timezone        VARCHAR             NOT NULL,
    creation_date   TIMESTAMP           NOT NULL,
    latitude        DOUBLE PRECISION,
    longitude       DOUBLE PRECISION
);

CREATE TABLE t_meters
(
    meter_id        VARCHAR PRIMARY KEY NOT NULL,
    installation_id VARCHAR             NOT NULL,
    meter_type      VARCHAR             NOT NULL,
    prim_ad         INTEGER             NOT NULL,
    virtual         BOOLEAN             NOT NULL,
    CONSTRAINT meters_installation_id
        FOREIGN KEY (installation_id)
            REFERENCES t_installations (installation_id)
);

CREATE TABLE t_installation_values
(
    installation_id VARCHAR          NOT NULL,
    date_time       TIMESTAMP        NOT NULL,
    prod_total      DOUBLE PRECISION NOT NULL,
    self            DOUBLE PRECISION NOT NULL,
    to_ext          DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (installation_id, date_time),
    CONSTRAINT installation_values_installation_id
        FOREIGN KEY (installation_id)
            REFERENCES t_installations (installation_id)
);

CREATE TABLE t_meter_values
(
    meter_id  VARCHAR          NOT NULL,
    date_time TIMESTAMP        NOT NULL,
    total     DOUBLE PRECISION NOT NULL,
    self      DOUBLE PRECISION NOT NULL,
    ext       DOUBLE PRECISION NOT NULL,
    PRIMARY KEY (meter_id, date_time),
    CONSTRAINT meter_values_meter_id
        FOREIGN KEY (meter_id)
            REFERENCES t_meters (meter_id)
);