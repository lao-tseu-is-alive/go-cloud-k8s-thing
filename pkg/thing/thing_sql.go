package thing

const (
	listThings = `SELECT 
		id,
       type_id,
       name,
       description,
       external_id,
       inactivated,
       _created_by as created_by,
       _created_at as created_at 
FROM go_thing.thing
WHERE _deleted = false
AND type_id = coalesce($3, type_id)
AND _created_by = coalesce($4, _created_by)
AND inactivated = coalesce($5, inactivated)
--AND validated = $6
ORDER BY _created_at DESC
LIMIT $1 OFFSET $2;
`
	typeThingCount = "SELECT COUNT(*) FROM go_thing.type_thing;"
	createThing    = `
INSERT INTO go_thing.thing
(id, type_id, name, description, comment, external_id, external_ref,
 build_at, status, contained_by, contained_by_old,validated, validated_time, validated_by,
 managed_by, _created_at, _created_by, more_data, text_search, position)
VALUES ($1, $2, $3, $4, $5, $6, $7,
        $8, $9, $10, $11, $12, $13, $14,
        $15, CURRENT_TIMESTAMP, $16, $17,
        to_tsvector('french', unaccent($3) ||
                              ' ' || coalesce(unaccent($4), ' ') ||
                              ' ' || coalesce(unaccent($5), ' ') ),
        ST_SetSRID(ST_MakePoint($18,$19), 2056));
`

	getThing = `SELECT id,
       type_id,
       name,
       description,
       comment,
       external_id,
       external_ref,
       build_at,
       status,
       contained_by,
       contained_by_old,
       inactivated,
       inactivated_time,
       inactivated_by,
       inactivated_reason,
       validated,
       validated_time,
       validated_by,
       managed_by,
       _created_at as create_at,
       _created_by as create_by,
       _last_modified_at as last_modified_at,
       _last_modified_by as last_modified_by,
       _deleted as deleted,
       _deleted_at as deleted_at,
       _deleted_by as deleted_by,
       more_data, 
       round(st_x(ST_Centroid(position))::numeric, 2) AS pos_x,
       round(st_y(ST_Centroid(position))::numeric, 2) AS pos_y
FROM go_thing.thing
WHERE id = $1;
`
	existThing        = `SELECT COUNT(*) FROM go_thing.thing WHERE id = $1;`
	isActiveThing     = `SELECT COUNT(*) FROM go_thing.thing WHERE inactivated=false AND id = $1;`
	existThingOwnedBy = `SELECT COUNT(*) FROM go_thing.thing WHERE id = $1 AND _created_by = $2;`
	countThing        = `SELECT COUNT(*) FROM go_thing.thing;`
	deleteThing       = `
UPDATE go_thing.thing
SET
    _deleted = true,
    _deleted_by = $1,
    _deleted_at = CURRENT_TIMESTAMP
WHERE id = $2;`
)
