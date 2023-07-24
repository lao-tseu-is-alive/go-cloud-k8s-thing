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
LIMIT $1;
`
	typeThingCount = "SELECT COUNT(*) FROM go_thing.type_thing;"
)
