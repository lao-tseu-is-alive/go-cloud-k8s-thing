package thing

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

// Storage is an interface to different implementation of persistence for Things/TypeThing
type Storage interface {
	// GeoJson returns a geoJson of existing things with the given offset and limit.
	GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error)
	// List returns the list of existing things with the given offset and limit.
	List(ctx context.Context, offset, limit int, params ListParams) ([]*ThingList, error)
	// ListByExternalId returns the list of existing things having the given externalId with the given offset and limit.
	ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*ThingList, error)
	// Search returns the list of existing things filtered by search params with the given offset and limit.
	Search(ctx context.Context, offset, limit int, params SearchParams) ([]*ThingList, error)
	// Get returns the thing with the specified things ID.
	Get(ctx context.Context, id uuid.UUID) (*Thing, error)
	// Exist returns true only if a things with the specified id exists in store.
	Exist(ctx context.Context, id uuid.UUID) bool
	// Count returns the total number of things.
	Count(ctx context.Context, params CountParams) (int32, error)
	// Create saves a new things in the storage.
	Create(ctx context.Context, thing Thing) (*Thing, error)
	// Update updates the things with given ID in the storage.
	Update(ctx context.Context, id uuid.UUID, thing Thing) (*Thing, error)
	// Delete removes the things with given ID from the storage.
	Delete(ctx context.Context, id uuid.UUID, userId int32) error
	// IsThingActive returns true if the thing with the specified id has the inactivated attribute set to false
	IsThingActive(ctx context.Context, id uuid.UUID) bool
	// IsUserOwner returns true only if userId is the creator of the record (owner) of this thing in store.
	IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool
	// CreateTypeThing saves a new typeThing in the storage.
	CreateTypeThing(ctx context.Context, typeThing TypeThing) (*TypeThing, error)
	// UpdateTypeThing updates the typeThing with given ID in the storage.
	UpdateTypeThing(ctx context.Context, id int32, typeThing TypeThing) (*TypeThing, error)
	// DeleteTypeThing removes the typeThing with given ID from the storage.
	DeleteTypeThing(ctx context.Context, id int32, userId int32) error
	// ListTypeThing returns the list of active typeThings with the given offset and limit.
	ListTypeThing(ctx context.Context, offset, limit int, params TypeThingListParams) ([]*TypeThingList, error)
	// GetTypeThing returns the typeThing with the specified things ID.
	GetTypeThing(ctx context.Context, id int32) (*TypeThing, error)
	// CountTypeThing returns the number of TypeThing based on search criteria
	CountTypeThing(ctx context.Context, params TypeThingCountParams) (int32, error)
}

func GetStorageInstanceOrPanic(dbDriver string, db database.DB, l *slog.Logger) Storage {
	var store Storage
	var err error
	switch dbDriver {
	case "pgx":
		store, err = NewPgxDB(db, l)
		if err != nil {
			l.Error("error doing NewPgxDB", "error", err)
			panic(err)
		}

	default:
		panic("unsupported DB driver type")
	}
	return store
}
