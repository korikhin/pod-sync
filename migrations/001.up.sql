create schema if not exists watcher;

create table if not exists watcher.clients (
    id serial primary key,
    name varchar(50) not null,
    version smallint not null,
    image varchar(50) not null,
    cpu varchar(10) not null,
    mem varchar(10) not null,
    priority real not null default 0,
    created_at timestamp not null default timezone('UTC', now()),
    updated_at timestamp not null default timezone('UTC', now())
);

create table if not exists watcher.status (
    id serial primary key,
    client_id integer not null unique,
    "X" bool not null default false,
    "Y" bool not null default false,
    "Z" bool not null default false,

    foreign key (client_id) references watcher.clients (id) on delete cascade
);
