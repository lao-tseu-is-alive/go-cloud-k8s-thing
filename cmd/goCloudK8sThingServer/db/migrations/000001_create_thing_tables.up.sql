-- CREATE EXTENSION postgis; this one should be done by a superuser of db
/*
-- Table structure for table thing for project go-cloud-k8s-thing
-- version : 0.0.3
*/
CREATE TABLE IF NOT EXISTS public.thing
(
    -- using Postgres Native UUID v4 128bit https://www.postgresql.org/docs/14/datatype-uuid.html
    -- this choice allows to do client side generation of the id UUID v4
    id                 uuid    not null
        constraint pk_thing primary key default gen_random_uuid(),
    type_id            integer not null,
    name               text    not null,
    description        text,
    comment            text,
    external_id        integer,
    external_ref       text,
    build_time         date,
    contained_by       integer,
    inactivated        boolean          default false,
    inactivated_time   timestamp,
    inactivated_by     integer,
    inactivated_reason text,
    validated          boolean          default false,
    validated_time     timestamp,
    validated_by       integer,
    created_at         timestamp        default now() not null,
    created_by         integer not null,
    last_modified_at   timestamp,
    last_modified_by   integer,
    deleted            boolean          default false,
    deleted_at         timestamp,
    deleted_by         integer,
    additional_data    jsonb,
    text_search        tsvector
);
alter table public.thing
    owner to go_cloud_k8s_thing;

SELECT AddGeometryColumn('thing', 'geom', 2056, 'POINT', 2);
CREATE INDEX idx_thing_geom_gist ON thing USING gist (geom);
CREATE INDEX idx_thing_type_id ON thing (type_id);



--
-- Table structure for table `TypeThing` generated from model 'TypeThing'
--

CREATE TABLE IF NOT EXISTS public.type_thing
(
    id                 serial
        constraint pk_type_thing primary key,
    name               text                    not null,
    description        text,
    comment            text,
    external_id        integer,
    table_name         text,
    inactivated        boolean   default false,
    inactivated_time   timestamp,
    inactivated_by     integer,
    inactivated_reason text,
    validated_time     timestamp,
    managed_by_group   integer,
    created_at         timestamp default now() not null,
    created_by         integer                 not null,
    last_modified_at   timestamp,
    last_modified_by   integer,
    deleted            boolean   default false,
    deleted_at         timestamp,
    deleted_by         integer
);

alter table public.type_thing
    owner to go_cloud_k8s_thing;

