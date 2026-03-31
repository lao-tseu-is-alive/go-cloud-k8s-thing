package thing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
)

// MaxPaginationLimit defines the maximum number of items that can be requested in a single list/search API call
const MaxPaginationLimit = 1000

// BusinessService Business Service contains the transport-agnostic business logic for Thing operations
type BusinessService struct {
	Log              *slog.Logger
	DbConn           database.DB
	Store            Storage
	ListDefaultLimit int
}

// NewBusinessService creates a new instance of BusinessService
func NewBusinessService(store Storage, dbConn database.DB, log *slog.Logger, listDefaultLimit int) *BusinessService {
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
func (s *BusinessService) GeoJson(ctx context.Context, req *thingv1.GeoJsonRequest) (string, error) {
	if req.Limit > MaxPaginationLimit {
		req.Limit = MaxPaginationLimit
	}
	jsonResult, err := s.Store.GeoJson(ctx, req)
	if err != nil {
		if errors.Is(err, ErrEmptyResult) {
			return EmptyGeoJson, nil
		}
		return "", fmt.Errorf("error retrieving geoJson: %w", err)
	}
	if jsonResult == "" {
		return EmptyGeoJson, nil
	}
	return jsonResult, nil
}

// List returns the list of things based on the given parameters
func (s *BusinessService) List(ctx context.Context, req *thingv1.ListRequest) ([]*thingv1.ThingList, error) {
	if req.Limit > MaxPaginationLimit {
		req.Limit = MaxPaginationLimit
	}
	list, err := s.Store.List(ctx, req)
	if err != nil {
		if errors.Is(err, ErrEmptyResult) {
			// No rows is not an error, return empty slice
			return make([]*thingv1.ThingList, 0), nil
		}
		return nil, fmt.Errorf("error listing things: %w", err)
	}
	if list == nil {
		return make([]*thingv1.ThingList, 0), nil
	}
	return list, nil
}

// Create creates a new thing with the given data
func (s *BusinessService) Create(ctx context.Context, currentUserId int32, newThing *thingv1.Thing) (*thingv1.Thing, error) {
	// Validate name
	if err := validateName(newThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeThingCount, err := s.DbConn.GetQueryInt(ctx, existTypeThing, newThing.TypeId)
	if err != nil || typeThingCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeThingNotFound, newThing.TypeId)
	}

	// Check if thing already exists
	thingUUID, err := uuid.Parse(newThing.Id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid uuid %v", ErrInvalidInput, newThing.Id)
	}
	exists, err := s.Store.Exist(ctx, thingUUID)
	if err != nil {
		return nil, fmt.Errorf("error verifying existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("%w: id %v", ErrAlreadyExists, newThing.Id)
	}

	// Set creator
	newThing.CreatedBy = currentUserId

	// Create in storage
	thingCreated, err := s.Store.Create(ctx, newThing)
	if err != nil {
		return nil, fmt.Errorf("error creating thing: %w", err)
	}

	s.Log.Info("Created thing", "id", thingCreated.Id, "userId", currentUserId)
	return thingCreated, nil
}

// Count returns the number of things based on the given parameters
func (s *BusinessService) Count(ctx context.Context, req *thingv1.CountRequest) (int32, error) {
	numThings, err := s.Store.Count(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("error counting things: %w", err)
	}
	return numThings, nil
}

// Delete removes a thing with the given ID
func (s *BusinessService) Delete(ctx context.Context, currentUserId int32, thingId uuid.UUID) error {
	// Check if thing exists
	exists, err := s.Store.Exist(ctx, thingId)
	if err != nil {
		return fmt.Errorf("error verifying existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: id %v", ErrNotFound, thingId)
	}

	// Check if user is owner
	isOwner, err := s.Store.IsUserOwner(ctx, thingId, currentUserId)
	if err != nil {
		return fmt.Errorf("error verifying ownership: %w", err)
	}
	if !isOwner {
		return fmt.Errorf("%w: user %d is not owner of thing %v", ErrUnauthorized, currentUserId, thingId)
	}

	// Delete from storage
	err = s.Store.Delete(ctx, thingId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting thing: %w", err)
	}

	s.Log.Info("Deleted thing", "id", thingId, "userId", currentUserId)
	return nil
}

// Get retrieves a thing by its ID
func (s *BusinessService) Get(ctx context.Context, thingId uuid.UUID) (*thingv1.Thing, error) {
	// Check if thing exists
	exists, err := s.Store.Exist(ctx, thingId)
	if err != nil {
		return nil, fmt.Errorf("error verifying existence: %w", err)
	}
	if !exists {
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
func (s *BusinessService) Update(ctx context.Context, currentUserId int32, thingId uuid.UUID, updateThing *thingv1.Thing) (*thingv1.Thing, error) {
	// Check if thing exists
	exists, err := s.Store.Exist(ctx, thingId)
	if err != nil {
		return nil, fmt.Errorf("error verifying existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("%w: id %v", ErrNotFound, thingId)
	}

	// Check if user is owner
	isOwner, err := s.Store.IsUserOwner(ctx, thingId, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("error verifying ownership: %w", err)
	}
	if !isOwner {
		return nil, fmt.Errorf("%w: user %d is not owner of thing %v", ErrUnauthorized, currentUserId, thingId)
	}

	// Validate name
	if err := validateName(updateThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Validate TypeId
	typeThingCount, err := s.DbConn.GetQueryInt(ctx, existTypeThing, updateThing.TypeId)
	if err != nil || typeThingCount < 1 {
		return nil, fmt.Errorf("%w: typeId %v", ErrTypeThingNotFound, updateThing.TypeId)
	}

	// Set last modifier
	updateThing.LastModifiedBy = currentUserId

	// Update in storage
	thingUpdated, err := s.Store.Update(ctx, thingId, updateThing)
	if err != nil {
		return nil, fmt.Errorf("error updating thing: %w", err)
	}

	s.Log.Info("Updated thing", "id", thingId, "userId", currentUserId)
	return thingUpdated, nil
}

// ListByExternalId returns things filtered by external ID
func (s *BusinessService) ListByExternalId(ctx context.Context, req *thingv1.ListByExternalIdRequest) ([]*thingv1.ThingList, error) {
	if req.Limit > MaxPaginationLimit {
		req.Limit = MaxPaginationLimit
	}
	list, err := s.Store.ListByExternalId(ctx, req)
	if err != nil {
		if errors.Is(err, ErrEmptyResult) {
			// No rows is not an error, return empty slice
			return make([]*thingv1.ThingList, 0), nil
		}
		return nil, fmt.Errorf("error listing things by external id: %w", err)
	}
	if list == nil {
		return make([]*thingv1.ThingList, 0), nil
	}
	return list, nil
}

// Search returns things based on search criteria
func (s *BusinessService) Search(ctx context.Context, req *thingv1.SearchRequest) ([]*thingv1.ThingList, error) {
	if req.Limit > MaxPaginationLimit {
		req.Limit = MaxPaginationLimit
	}
	list, err := s.Store.Search(ctx, req)
	if err != nil {
		if errors.Is(err, ErrEmptyResult) {
			// No rows is not an error, return empty slice
			return make([]*thingv1.ThingList, 0), nil
		}
		return nil, fmt.Errorf("error searching things: %w", err)
	}
	if list == nil {
		return make([]*thingv1.ThingList, 0), nil
	}
	return list, nil
}

// ListTypeThings returns a list of TypeThing based on parameters
func (s *BusinessService) ListTypeThings(ctx context.Context, req *thingv1.TypeThingServiceListRequest) ([]*thingv1.TypeThingList, error) {
	if req.Limit > MaxPaginationLimit {
		req.Limit = MaxPaginationLimit
	}
	list, err := s.Store.ListTypeThing(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error listing type things: %w", err)
	}
	if list == nil {
		return make([]*thingv1.TypeThingList, 0), nil
	}
	return list, nil
}

// CreateTypeThing creates a new TypeThing
func (s *BusinessService) CreateTypeThing(ctx context.Context, currentUserId int32, isAdmin bool, newTypeThing *thingv1.TypeThing) (*thingv1.TypeThing, error) {
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

	s.Log.Info("Created TypeThing", "id", typeThingCreated.Id, "userId", currentUserId)
	return typeThingCreated, nil
}

// CountTypeThings returns the count of TypeThings based on parameters
func (s *BusinessService) CountTypeThings(ctx context.Context, req *thingv1.TypeThingServiceCountRequest) (int32, error) {
	numThings, err := s.Store.CountTypeThing(ctx, req)
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
	typeThingCount, err := s.DbConn.GetQueryInt(ctx, existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		return fmt.Errorf("%w: id %d", ErrTypeThingNotFound, typeThingId)
	}

	// Delete from storage
	err = s.Store.DeleteTypeThing(ctx, typeThingId, currentUserId)
	if err != nil {
		return fmt.Errorf("error deleting type thing: %w", err)
	}

	s.Log.Info("Deleted TypeThing", "id", typeThingId, "userId", currentUserId)
	return nil
}

// GetTypeThing retrieves a TypeThing by ID
func (s *BusinessService) GetTypeThing(ctx context.Context, isAdmin bool, typeThingId int32) (*thingv1.TypeThing, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeThing exists
	typeThingCount, err := s.DbConn.GetQueryInt(ctx, existTypeThing, typeThingId)
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
func (s *BusinessService) UpdateTypeThing(ctx context.Context, currentUserId int32, isAdmin bool, typeThingId int32, updateTypeThing *thingv1.TypeThing) (*thingv1.TypeThing, error) {
	// Check admin privileges
	if !isAdmin {
		return nil, ErrAdminRequired
	}

	// Check if TypeThing exists
	typeThingCount, err := s.DbConn.GetQueryInt(ctx, existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		return nil, fmt.Errorf("%w: id %d", ErrTypeThingNotFound, typeThingId)
	}

	// Validate name
	if err := validateName(updateTypeThing.Name); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Set last modifier
	updateTypeThing.LastModifiedBy = currentUserId

	// Update in storage
	thingUpdated, err := s.Store.UpdateTypeThing(ctx, typeThingId, updateTypeThing)
	if err != nil {
		return nil, fmt.Errorf("error updating type thing: %w", err)
	}

	s.Log.Info("Updated TypeThing", "id", typeThingId, "userId", currentUserId)
	return thingUpdated, nil
}
