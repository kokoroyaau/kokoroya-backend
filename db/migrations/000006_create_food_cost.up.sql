create table suppliers (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    name text not null,
    sort_order int not null default 0,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
create index suppliers_branch_idx on suppliers (branch_id);

create table purchase_entries (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    supplier_id bigint not null references suppliers(id) on delete cascade,
    purchase_date date not null,
    amount numeric(15,2) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (supplier_id, purchase_date)
);
create index purchase_entries_branch_date_idx on purchase_entries (branch_id, purchase_date);

create table gross_sales_entries (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    sales_date date not null,
    amount numeric(15,2) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (branch_id, sales_date)
);

create table weekly_net_sales_rates (
    id bigserial primary key,
    branch_id bigint not null references branches(id) on delete cascade,
    week_start_date date not null,
    rate numeric(6,4) not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (branch_id, week_start_date)
);
