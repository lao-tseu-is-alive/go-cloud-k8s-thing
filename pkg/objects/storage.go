package objects

import (
	"errors"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"log"
)

// Storage is an interface to different implementation of persistence for Objects/TypeObject
type Storage interface {
	// List returns the list of existing objects with the given offset and limit.
	List(offset, limit int) ([]*ObjectList, error)
	// Get returns the object with the specified objects ID.
	Get(id int32) (*Object, error)
	// GetMaxId returns the maximum value of objects id existing in store.
	GetMaxId() (int32, error)
	// Exist returns true only if a objects with the specified id exists in store.
	Exist(id int32) bool
	// Count returns the total number of objects.
	Count() (int32, error)
	// Create saves a new objects in the storage.
	Create(object Object) (*Object, error)
	// Update updates the objects with given ID in the storage.
	Update(id int32, object Object) (*Object, error)
	// Delete removes the objects with given ID from the storage.
	Delete(id int32) error
	// SearchObjectsByName list of existing objects where the name contains the given search pattern or err if not found
	SearchObjectsByName(pattern string) ([]*ObjectList, error)
	// IsObjectActive returns true if the object with the specified id has the is_active attribute set to true
	IsObjectActive(id int32) bool
	// CreateTypeObject saves a new typeObject in the storage.
	CreateTypeObject(typeObject TypeObject) (*TypeObject, error)
	// UpdateTypeObject updates the typeObject with given ID in the storage.
	UpdateTypeObject(id int32, typeObject TypeObject) (*TypeObject, error)
	// DeleteTypeObject removes the typeObject with given ID from the storage.
	DeleteTypeObject(id int32) error
	// ListTypeObject returns the list of active typeObjects with the given offset and limit.
	ListTypeObject(offset, limit int) ([]*TypeObjectList, error)
	// GetTypeObject returns the typeObject with the specified objects ID.
	GetTypeObject(id int32) (*TypeObject, error)
}

func GetStorageInstance(dbDriver string, db database.DB, l *log.Logger) (Storage, error) {
	var store Storage
	switch dbDriver {
	case "pgx":
		pgConn, err := db.GetPGConn()
		if err != nil {
			return nil, err
		}
		store = PGX{
			con: pgConn,
			log: l,
		}

	default:
		return nil, errors.New("unsupported DB driver type")
	}
	return store, nil
}
