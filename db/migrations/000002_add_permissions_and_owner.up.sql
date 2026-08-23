alter table users add column if not exists permissions text[] not null default '{}';
