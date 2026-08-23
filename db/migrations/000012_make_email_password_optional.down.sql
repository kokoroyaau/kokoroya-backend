alter table users
    alter column email set not null,
    alter column password_hash set not null;
