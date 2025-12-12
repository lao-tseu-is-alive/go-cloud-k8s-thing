package thing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

// BusinessService Business Service contains the transport-agnostic business logic for Thing operations
type BusinessService struct {
	Log              golog.MyLogger
	DbConn           database.DB
	Store            Storage
	ListDefaultLimit int
}

// NewBusinessService creates a new instance of BusinessService
func NewBusinessService(store Storage, dbConn database.DB, log golog.MyLogger, listDefaultLimit int) *BusinessService {
	return &BusinessService{
		Log:              log,
		DbConn:           dbConn,
		Store:            store,
		ListDefaultLimit: listDefaultLimit,
	}
}

// validateName validates the name field according to business rules
func validateName(name string) error {
	if len(strings.Trim(name, " ")) < 1 {
		return fmt.Errorf(FieldCannotBeEmpty, "name")
	}
	if len(name) < MinNameLength {
		return fmt.Errorf(FieldMinLengthIsN, "name", MinNameLength)
	}
	return nil
}

// GeoJson returns a geoJson representation of things based on the given parameters
func (s *BusinessService) GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error) {
	jsonResult, err := s.Store.GeoJson(ctx, offset, limit, params)
	if err != nil {
		return "", fmt.Errorf("error retrieving geoJson: %w", err)
	}
	if jsonResult == "" {
		return "empty", nil
	}
	return jsonResult, nil
}

// List returns the list of things based on the given parameters
func (s *BusinessService) List(ctx context.Context, offset, limit int, params ListParams) ([]*ThingList, error) {
	list, err := s.Store.List(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*ThingList, 0), nil
		}
		return nil, fmt.Errorf("error listing things: %w", err)
	}
	if list == nil {
		return make([]*ThingList, 0), nil
	}
	return list, nil
}

// Create creates a new thing with the given data
func (s *BusinessService) Create(ctx context.Context, currentUserId int32, newThing Thing) (*Thing, error) {
	// Validate name
	if err := validateName(newThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Check if thing already exists
	if s.Store.Exist(ctx, newThing.Id) {
		return nil, fmt.Errorf("%w: id %v", ErrAlreadyExists, newThing.Id)
	}

	// Set creator
	newThing.CreatedBy = currentUserId

	// Create in storage
	thingCreated, err := s.Store.Create(ctx, newThing)
	if err != nil {
		return nil, fmt.Errorf("error creating thing: %w", err)
	}

	s.Log.Info("Created thing with id: %v by user: %d", thingCreated.Id, currentUserId)
	return thingCreated, nil
}

// Count returns the number of things based on the given parameters
func (s *BusinessService) Count(ctx context.Context, params CountParams) (int32, error) {
	numThings, err := s.Store.Count(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting things: %w", err)
	}
	return numThings, nil
}

// Delete removes a thing with the given ID
func (s *BusinessService) Delete(ctx context.Context, currentUserId int32, thingId uuid.UUID) error {
	// Check if thing exists
	if !s.Store.Exist(ctx, thingId) {
		return fmt.Errorf("%w: id %v", ErrNotFound, thingId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, thingId, currentUserId) {
		return fmt.Errorf("%w: user %d is not owner of thing %v", ErrUnauthorized, currentUserId, thingId)
	}

	// Delete from storage
	err := s.Store.Delete(ctx, thingId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting thing: %w", err)
	}

	s.Log.Info("Deleted thing %v by user: %d", thingId, currentUserId)
	return nil
}

// Get retrieves a thing by its ID
func (s *BusinessService) Get(ctx context.Context, thingId uuid.UUID) (*Thing, error) {
	// Check if thing exists
	if !s.Store.Exist(ctx, thingId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, thingId)
	}

	// Get from storage
	thing, err := s.Store.Get(ctx, thingId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving thing: %w", err)
	}

	return thing, nil
}

// Update updates a thing with the given ID
func (s *BusinessService) Update(ctx context.Context, currentUserId int32, thingId uuid.UUID, updateThing Thing) (*Thing, error) {
	// Check if thing exists
	if !s.Store.Exist(ctx, thingId) {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, thingId)
	}

	// Check if user is owner
	if !s.Store.IsUserOwner(ctx, thingId, currentUserId) {
		return nil, fmt.Errorf("%w: user %d is not owner of thing %v", ErrUnauthorized, currentUserId, thingId)
	}

	// Validate name
	if err := validateName(updateThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set last modifier
	updateThing.LastModifiedBy = &currentUserId

	// Update in storage
	thingUpdated, err := s.Store.Update(ctx, thingId, updateThing)
	if err != nil {
		return nil, fmt.Errorf("error updating thing: %w", err)
	}

	s.Log.Info("Updated thing %v by user: %d", thingId, currentUserId)
	return thingUpdated, nil
}

// ListByExternalId returns things filtered by external ID
func (s *BusinessService) ListByExternalId(ctx context.Context, offset, limit, externalId int) ([]*ThingList, error) {
	list, err := s.Store.ListByExternalId(ctx, offset, limit, externalId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*ThingList, 0), nil
		}
		return nil, fmt.Errorf("error listing things by external id: %w", err)
	}
	if list == nil {
		return make([]*ThingList, 0), nil
	}
	return list, nil
}

// Search returns things based on search criteria
func (s *BusinessService) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*ThingList, error) {
	list, err := s.Store.Search(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No rows is not an error, return empty slice
			return make([]*ThingList, 0), nil
		}
		return nil, fmt.Errorf("error searching things: %w", err)
	}
	if list == nil {
		return make([]*ThingList, 0), nil
	}
	return list, nil
}

// ListTypeThings returns a list of TypeThing based on parameters
func (s *BusinessService) ListTypeThings(ctx context.Context, offset, limit int, params TypeThingListParams) ([]*TypeThingList, error) {
	list, err := s.Store.ListTypeThing(ctx, offset, limit, params)
	if err != nil {
		return nil, fmt.Errorf("error listing type things: %w", err)
	}
	if list == nil {
		return make([]*TypeThingList, 0), nil
	}
	return list, nil
}

// CreateTypeThing creates a new TypeThing
func (s *BusinessService) CreateTypeThing(ctx context.Context, currentUserId int32, isAdmin bool, newTypeThing TypeThing) (*TypeThing, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Validate name
	if err := validateName(newTypeThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set creator
	newTypeThing.CreatedBy = currentUserId

	// Create in storage
	typeThingCreated, err := s.Store.CreateTypeThing(ctx, newTypeThing)
	if err != nil {
		return nil, fmt.Errorf("error creating type thing: %w", err)
	}

	s.Log.Info("Created TypeThing with id: %d by user: %d", typeThingCreated.Id, currentUserId)
	return typeThingCreated, nil
}

// CountTypeThings returns the count of TypeThings based on parameters
func (s *BusinessService) CountTypeThings(ctx context.Context, params TypeThingCountParams) (int32, error) {
	numThings, err := s.Store.CountTypeThing(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error counting type things: %w", err)
	}
	return numThings, nil
}

// DeleteTypeThing deletes a TypeThing by ID
func (s *BusinessService) DeleteTypeThing(ctx context.Context, currentUserId int32, isAdmin bool, typeThingId int32) error {
	// Check admin privileges
	if !isAdmin {
		return ErrAdminRequired
	}

	// Check if TypeThing exists
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		return fmt.Errorf("%w: id %d", ErrTypeThingNotFound, typeThingId)
	}

	// Delete from storage
	err = s.Store.DeleteTypeThing(ctx, typeThingId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting type thing: %w", err)
	}

	s.Log.Info("Deleted TypeThing %d by user: %d", typeThingId, currentUserId)
	return nil
}

// GetTypeThing retrieves a TypeThing by ID
func (s *BusinessService) GetTypeThing(ctx context.Context, isAdmin bool, typeThingId int32) (*TypeThing, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeThing exists
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeThingNotFound, typeThingId)
	}

	// Get from storage
	typeThing, err := s.Store.GetTypeThing(ctx, typeThingId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving type thing: %w", err)
	}

	return typeThing, nil
}

// UpdateTypeThing updates a TypeThing
func (s *BusinessService) UpdateTypeThing(ctx context.Context, currentUserId int32, isAdmin bool, typeThingId int32, updateTypeThing TypeThing) (*TypeThing, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeThing exists
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeThingNotFound, typeThingId)
	}

	// Validate name
	if err := validateName(updateTypeThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set last modifier
	updateTypeThing.LastModifiedBy = &currentUserId

	// Update in storage
	thingUpdated, err := s.Store.UpdateTypeThing(ctx, typeThingId, updateTypeThing)
	if err != nil {
		return nil, fmt.Errorf("error updating type thing: %w", err)
	}

	s.Log.Info("Updated TypeThing %d by user: %d", typeThingId, currentUserId)
	return thingUpdated, nil
}
