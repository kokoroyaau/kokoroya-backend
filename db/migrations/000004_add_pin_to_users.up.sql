alter table users add column if not exists pin char(4);
create unique index if not exists users_pin_unique on users (pin) where pin is not null;
