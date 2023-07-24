SELECT COUNT(*), T.idtypething,
       (SELECT name FROM type_thing tt WHERE tt.id = T.idtypething)
FROM goeland_thing_gc T WHERE idcreator is null
GROUP BY idtypething;


with old as (select   * from goeland_thing_gc)
INSERT INTO thing
(type_id, name, description,  external_id,
 build_time, contained_by, inactivated,
 validated, validated_time, created_at, created_by,
 last_modified_at, last_modified_by, position)
SELECT
    old.idtypething, old.name, old.description, old.idthing,
    old.dateconstruction,old.idcontainer, not(old.isactive),
    old.isvalidated, old.datevalidation, old.datecreated, coalesce(old.idcreator, 1),
    old.datelastmodif, old.idmodificator, old.thing_position
FROM old;

-- to_tsvector('french', coalesce(unaccent(name), ''), )

select count(*) FROM goeland_thing_gc;
select count(*) FROM thing;
