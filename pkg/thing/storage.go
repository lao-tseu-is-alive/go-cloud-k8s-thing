package thing

import (
	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

// Storage is an interface to different implementation of persistence for Things/TypeThing
type Storage interface {
	// GeoJson returns a geoJson of existing things with the given offset and limit.
	GeoJson(offset, limit int, params GeoJsonParams) (string, error)
	// List returns the list of existing things with the given offset and limit.
	List(offset, limit int, params ListParams) ([]*ThingList, error)
	// ListByExternalId returns the list of existing things having the given externalId with the given offset and limit.
	ListByExternalId(offset, limit int, externalId int) ([]*ThingList, error)
	// Search returns the list of existing things filtered by search params with the given offset and limit.
	Search(offset, limit int, params SearchParams) ([]*ThingList, error)
	// Get returns the thing with the specified things ID.
	Get(id uuid.UUID) (*Thing, error)
	// Exist returns true only if a things with the specified id exists in store.
	Exist(id uuid.UUID) bool
	// Count returns the total number of things.
	Count(params CountParams) (int32, error)
	// Create saves a new things in the storage.
	Create(thing Thing) (*Thing, error)
	// Update updates the things with given ID in the storage.
	Update(id uuid.UUID, thing Thing) (*Thing, error)
	// Delete removes the things with given ID from the storage.
	Delete(id uuid.UUID, userId int32) error
	// IsThingActive returns true if the thing with the specified id has the inactivated attribute set to false
	IsThingActive(id uuid.UUID) bool
	// IsUserOwner returns true only if userId is the creator of the record (owner) of this thing in store.
	IsUserOwner(id uuid.UUID, userId int32) bool
	// CreateTypeThing saves a new typeThing in the storage.
	CreateTypeThing(typeThing TypeThing) (*TypeThing, error)
	// UpdateTypeThing updates the typeThing with given ID in the storage.
	UpdateTypeThing(id int32, typeThing TypeThing) (*TypeThing, error)
	// DeleteTypeThing removes the typeThing with given ID from the storage.
	DeleteTypeThing(id int32, userId int32) error
	// ListTypeThing returns the list of active typeThings with the given offset and limit.
	ListTypeThing(offset, limit int, params TypeThingListParams) ([]*TypeThingList, error)
	// GetTypeThing returns the typeThing with the specified things ID.
	GetTypeThing(id int32) (*TypeThing, error)
	// CountTypeThing returns the number of TypeThing based on search criteria
	CountTypeThing(params TypeThingCountParams) (int32, error)
}

func GetStorageInstanceOrPanic(dbDriver string, db database.DB, l golog.MyLogger) Storage {
	var store Storage
	var err error
	switch dbDriver {
	case "pgx":
		store, err = NewPgxDB(db, l)
		if err != nil {
			l.Fatal("error doing NewPgxDB(pgConn : %w", err)
		}

	default:
		panic("unsupported DB driver type")
	}
	return store
}
