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
ORDER BY external_id
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
		to_tsvector('french', unaccent(name) ||
        ' ' || coalesce(unaccent(description), ' ') ||
        ' ' || coalesce(unaccent(comment), ' ') ), 
        ST_SetSRID(ST_MakePoint($18,$19), 2056));
`
)
