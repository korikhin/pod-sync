create schema if not exists vortex;

create table if not exists vortex.clients (
    id serial primary key,
    name varchar(50) not null,
    version smallint not null,
    image varchar(50) not null,
    cpu varchar(10) not null,
    mem varchar(10) not null,
    priority real not null default 0,
    spawned_at timestamp not null default timezone('UTC', now()), -- ?
    created_at timestamp not null default timezone('UTC', now()),
    updated_at timestamp not null default timezone('UTC', now())
);

create table if not exists vortex.status (
    id serial primary key,
    client_id integer not null unique,
    "VWAP" bool not null default false,
    "TWAP" bool not null default false,
    "HFT" bool not null default false,

    foreign key (client_id) references vortex.clients (id) on delete cascade
);
