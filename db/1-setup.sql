/* last run on postgres 12.3 */

create extension if not exists "uuid-ossp";

create table if not exists "channel_info" (
  "channel_id"  varchar primary key,
  "name"        varchar,
  "users"       varchar[],
  "updated"     timestamptz not null default now(),
  "expires"     timestamptz not null default now() + interval '1 day'
);

create table if not exists "feedback" (
  "id"          uuid primary key default uuid_generate_v4(),
  "user_id"     varchar not null,
  "channel_id"  varchar not null,
  "tag"         text,
  "content"     text,
  "created"     timestamptz not null default now()
);

create table if not exists "identity" (
  "username"    varchar primary key,
  "user_id"     varchar not null,
  "channel_id"  varchar not null,
  "created"     timestamptz not null default now()
);
