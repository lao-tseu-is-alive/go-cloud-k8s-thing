WITH swiss as(
SELECT id,lower(replace(replace(uuid,'{',''), '}', '')) as uuid,geom,name FROM swisstopo.communes sc
WHERE  sc.kantonsnum = 22
AND sc.name IN (SELECT name FROM go_thing.thing WHERE type_id = 2 AND _created_by != 999)
ORDER BY name)
UPDATE go_thing.thing
SET id=swiss.uuid::uuid, position=ST_PointOnSurface(swiss.geom)
FROM swiss
WHERE type_id = 2 AND _created_by != 999 AND go_thing.thing.name=swiss.name;


-- du coup on a une jointure possible pour une mise à jour ultérieure des données
SELECT t.id,t.name
FROM go_thing.thing t
INNER JOIN swisstopo.communes c ON c.uuid::uuid=t.id
