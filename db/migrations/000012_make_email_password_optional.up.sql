alter table users
    alter column email drop not null,
    alter column password_hash drop not null;
