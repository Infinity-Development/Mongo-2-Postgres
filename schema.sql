create table backups (col text not null, data jsonb not null, ts timestamptz not null default now());
