CREATE TABLE companies (
    name TEXT NOT NULL,
    site TEXT NOT NULL,

    CONSTRAINT name UNIQUE (
        name, site
    )
);
