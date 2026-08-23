update users set pin = null where pin is not null and length(trim(pin)) <> 4;
alter table users alter column pin type char(4);
