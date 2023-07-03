CREATE TABLE users (
    id bigserial primary key,
    name text not null,
    password text not null,
    created timestamptz default current_timestamp,
    updated timestamptz default current_timestamp
)