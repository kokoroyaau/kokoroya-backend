alter table users add column if not exists hour_cap_weekday numeric(6, 2);
alter table users add column if not exists hour_cap_weekend numeric(6, 2);
