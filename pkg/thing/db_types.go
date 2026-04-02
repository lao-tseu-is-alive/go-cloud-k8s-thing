package thing

import "time"

// ThingDB defines the database row representation for a Thing.
type ThingDB struct {
	Id                string                  `db:"id"` // UUID as string in Go for pgx mapping? Actually pgxscan needs string or uuid.UUID. The current was openapi_types.UUID
	TypeId            int32                   `db:"type_id"`
	Name              string                  `db:"name"`
	Description       *string                 `db:"description"`
	Comment           *string                 `db:"comment"`
	ExternalId        *int32                  `db:"external_id"`
	ExternalRef       *string                 `db:"external_ref"`
	BuildAt           *time.Time              `db:"build_at"`
	Status            *string                 `db:"status"`
	ContainedBy       *string                 `db:"contained_by"`
	ContainedByOld    *int32                  `db:"contained_by_old"`
	Inactivated       bool                    `db:"inactivated"`
	InactivatedTime   *time.Time              `db:"inactivated_time"`
	InactivatedBy     *int32                  `db:"inactivated_by"`
	InactivatedReason *string                 `db:"inactivated_reason"`
	Validated         *bool                   `db:"validated"`
	ValidatedTime     *time.Time              `db:"validated_time"`
	ValidatedBy       *int32                  `db:"validated_by"`
	ManagedBy         *int32                  `db:"managed_by"`
	CreatedAt         *time.Time              `db:"created_at"`
	CreatedBy         int32                   `db:"created_by"`
	LastModifiedAt    *time.Time              `db:"last_modified_at"`
	LastModifiedBy    *int32                  `db:"last_modified_by"`
	Deleted           bool                    `db:"deleted"`
	DeletedAt         *time.Time              `db:"deleted_at"`
	DeletedBy         *int32                  `db:"deleted_by"`
	MoreData          *map[string]interface{} `db:"more_data"`
	PosX              float64                 `db:"pos_x"`
	PosY              float64                 `db:"pos_y"`
}

// ThingListDB defines a light version for List queries.
type ThingListDB struct {
	Id          string     `db:"id"`
	TypeId      int32      `db:"type_id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	ExternalId  *int32     `db:"external_id"`
	Inactivated bool       `db:"inactivated"`
	Validated   *bool      `db:"validated"`
	Status      *string    `db:"status"`
	CreatedBy   int32      `db:"created_by"`
	CreatedAt   *time.Time `db:"created_at"`
	PosX        float64    `db:"pos_x"`
	PosY        float64    `db:"pos_y"`
}

// TypeThingDB defines the database row representation for a TypeThing.
type TypeThingDB struct {
	Id                int32                   `db:"id"`
	Name              string                  `db:"name"`
	Description       *string                 `db:"description"`
	Comment           *string                 `db:"comment"`
	ExternalId        *int32                  `db:"external_id"`
	TableName         *string                 `db:"table_name"`
	GeometryType      *string                 `db:"geometry_type"`
	Inactivated       bool                    `db:"inactivated"`
	InactivatedTime   *time.Time              `db:"inactivated_time"`
	InactivatedBy     *int32                  `db:"inactivated_by"`
	InactivatedReason *string                 `db:"inactivated_reason"`
	ManagedBy         *int32                  `db:"managed_by"`
	IconPath          string                  `db:"icon_path"`
	CreatedAt         *time.Time              `db:"created_at"`
	CreatedBy         int32                   `db:"created_by"`
	LastModifiedAt    *time.Time              `db:"last_modified_at"`
	LastModifiedBy    *int32                  `db:"last_modified_by"`
	Deleted           bool                    `db:"deleted"`
	DeletedAt         *time.Time              `db:"deleted_at"`
	DeletedBy         *int32                  `db:"deleted_by"`
	MoreDataSchema    *map[string]interface{} `db:"more_data_schema"`
}

// TypeThingListDB defines a light version for List queries.
type TypeThingListDB struct {
	Id           int32     `db:"id"`
	Name         string    `db:"name"`
	ExternalId   *int32    `db:"external_id"`
	IconPath     string    `db:"icon_path"`
	CreatedAt    time.Time `db:"created_at"`
	TableName    *string   `db:"table_name"`
	GeometryType *string   `db:"geometry_type"`
	Inactivated  bool      `db:"inactivated"`
}
