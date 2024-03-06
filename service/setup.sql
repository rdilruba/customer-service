create table public.customers
(
    id      serial
        primary key,
    name    varchar(255),
    email   varchar(255)
        unique,
    address varchar(255)
);

alter table public.customers
    owner to postgres;

