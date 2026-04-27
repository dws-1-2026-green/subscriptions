create extension if not exists pgcrypto;

create table
    if not exists subscriptions (
        id uuid primary key default gen_random_uuid (),
        source text not null,
        event_type text not null,
        target_url text not null,
        http_method text not null default 'POST' check (http_method in ('POST', 'PUT', 'PATCH')),
        headers jsonb not null default '{}'::jsonb,
        enabled boolean not null default true,
        created_at timestamptz not null default now ()
    );

create index if not exists idx_subscriptions_lookup on subscriptions (source, event_type)
where
    enabled = true;