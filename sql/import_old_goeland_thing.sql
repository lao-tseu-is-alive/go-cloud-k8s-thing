SELECT COUNT(*),
       T.idtypething,
       (SELECT name FROM goeland.goeland_type_thing tt WHERE tt.idtypething = T.idtypething)
FROM goeland.goeland_thing_gc T
WHERE idcreator is null
GROUP BY idtypething;

SELECT COUNT(*),
       t.idtypething,
       (SELECT name FROM goeland.goeland_type_thing tt WHERE tt.idtypething = t.idtypething),
       t.name,
       min(t.idthing),
       max(t.idthing)
FROM goeland.goeland_thing_gc t
WHERE isactive = true
GROUP BY t.name, t.idtypething
HAVING COUNT(*) > 1
ORDER BY 1 DESC;


with old as (select * from goeland.goeland_thing_gc where isactive = true)
INSERT
INTO go_thing.thing
(type_id, name, description, external_id,
 build_at, contained_by_old, inactivated,
 validated, validated_time, _created_at, _created_by,
 _last_modified_at, _last_modified_by, position)
SELECT old.idtypething,
       old.name,
       old.description,
       old.idthing,
       old.dateconstruction,
       old.idcontainer,
       not (old.isactive),
       old.isvalidated,
       old.datevalidation,
       old.datecreated,
       coalesce(old.idcreator, 1),
       old.datelastmodif,
       old.idmodificator,
       old.thing_position
FROM old;

-- to_tsvector('french', coalesce(unaccent(name), ''), )

select count(*)
FROM goeland.goeland_thing_gc
where isactive = true;
select count(*)
FROM go_thing.thing;


UPDATE
    go_thing.thing
SET text_search = to_tsvector('french',
                              unaccent(name) ||
                              ' ' || coalesce(unaccent(description), ' ') ||
                              ' ' || coalesce(unaccent(comment), ' '))
WHERE text_search IS NULL;

create index thing_text_search_index
    on go_thing.thing using gin (text_search);