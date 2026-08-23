drop index if exists users_pin_unique;
alter table users drop column if exists pin;
