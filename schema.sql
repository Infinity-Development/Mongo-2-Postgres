CREATE EXTENSION "uuid-ossp";
create table backups (col text not null, data jsonb not null, ts timestamptz not null default now(), id uuid not null default uuid_generate_v4());
