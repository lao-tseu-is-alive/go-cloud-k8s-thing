package thing

const (
	typeThingListQuery = `
SELECT id,
    name,
    external_id,
    table_name,
    geometry_type,
    inactivated
FROM go_thing.type_thing
WHERE _deleted = false 
`
	typeThingListOrderBy     = " ORDER BY id DESC LIMIT $1 OFFSET $2;"
	listTypeThingsConditions = `
 AND text_search @@ plainto_tsquery('french', unaccent($3)
 AND _created_by = coalesce($4, _created_by)
 AND external_id = coalesce($5, external_id)
 AND inactivated = coalesce($6, inactivated)  
`
	typeThingCount  = "SELECT COUNT(*) FROM go_thing.type_thing;"
	createTypeThing = `
INSERT INTO go_thing.type_thing
    (name, description, comment, external_id, table_name, geometry_type,
     managed_by, _created_at, _created_by, more_data_schema, text_search)
VALUES ($1, $2, $3, $4, $5, $6,
        $7, CURRENT_TIMESTAMP, $8, $9,
        to_tsvector('french', unaccent($1) ||
                              ' ' || coalesce(unaccent($2), ' ') ||
                              ' ' || coalesce(unaccent($3), ' ') ))
RETURNING id;
`

	getTypeThing = `
SELECT id,
       name,
       description,
       comment,
       external_id,
       table_name,
       geometry_type,
       inactivated,
       inactivated_time,
       inactivated_by,
       inactivated_reason,
       managed_by,
       _created_at as create_at,
       _created_by as create_by,
       _last_modified_at as last_modified_at,
       _last_modified_by as last_modified_by,
       _deleted as deleted,
       _deleted_at as deleted_at,
       _deleted_by as deleted_by,
       more_data_schema
FROM go_thing.type_thing
WHERE id = $1;
`
	existTypeThing        = `SELECT COUNT(*) FROM go_thing.type_thing WHERE id = $1 AND  _deleted = false;`
	isActiveTypeThing     = `SELECT COUNT(*) FROM go_thing.type_thing WHERE inactivated=false AND id = $1;`
	existTypeThingOwnedBy = `SELECT COUNT(*) FROM go_thing.type_thing WHERE id = $1 AND _created_by = $2;`
	countTypeThing        = `SELECT COUNT(*) FROM go_thing.type_thing;`
	deleteTypeThing       = `
UPDATE go_thing.type_thing
SET
    _deleted = true,
    _deleted_by = $1,
    _deleted_at = CURRENT_TIMESTAMP
WHERE id = $2;`
	updateTypeTing = `
UPDATE go_thing.type_thing
SET
    name               = $2,
    description        = $3,
    comment            = $4,
    external_id        = $5,
    table_name         = $6,
    geometry_type      = $7,
    inactivated        = $8,
    inactivated_time   = $9,
    inactivated_by     = $10,
    inactivated_reason = $11,
    managed_by         = $12,
    _last_modified_at  = CURRENT_TIMESTAMP,
    _last_modified_by  = $13,
    more_data_schema   = $14,
    text_search = to_tsvector('french', unaccent($2) ||
                             ' ' || coalesce(unaccent($3), ' ') ||
                             ' ' || coalesce(unaccent($4), ' ') )
WHERE id = $1;
`
)
