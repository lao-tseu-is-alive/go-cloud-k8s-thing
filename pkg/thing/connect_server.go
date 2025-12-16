// Package thing provides Connect RPC handlers for the ThingService.
package thing

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1/thingv1connect"
)

// ThingConnectServer implements the ThingServiceHandler interface.
// Authentication is handled by the AuthInterceptor, which injects user info into context.
type ThingConnectServer struct {
	BusinessService *BusinessService
	Log             *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	thingv1connect.UnimplementedThingServiceHandler
}

// NewThingConnectServer creates a new ThingConnectServer.
// Note: Authentication is handled by the AuthInterceptor, not by this server.
func NewThingConnectServer(business *BusinessService, log *slog.Logger) *ThingConnectServer {
	return &ThingConnectServer{
		BusinessService: business,
		Log:             log,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// mapErrorToConnect converts business errors to Connect errors
func (s *ThingConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrTypeThingNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrNotOwner):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrAdminRequired):
		return connect.NewError(connect.CodePermissionDenied, errors.New(OnlyAdminCanManageTypeThings))
	case errors.Is(err, ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, pgx.ErrNoRows):
		return connect.NewError(connect.CodeNotFound, errors.New("not found"))
	default:
		s.Log.Error("internal error", "error", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}

// =============================================================================
// ThingService RPC Methods
// =============================================================================

// List returns a list of things
func (s *ThingConnectServer) List(
	ctx context.Context,
	req *connect.Request[thingv1.ListRequest],
) (*connect.Response[thingv1.ListResponse], error) {
	s.Log.Info("Connect: List called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("List", "userId", userId)

	// Build domain params from proto request
	msg := req.Msg
	params := ListParams{}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	// Handle pagination with defaults
	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	// Call business logic
	list, err := s.BusinessService.List(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Convert to proto and return
	response := &thingv1.ListResponse{
		Things: DomainThingListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new thing
func (s *ThingConnectServer) Create(
	ctx context.Context,
	req *connect.Request[thingv1.CreateRequest],
) (*connect.Response[thingv1.CreateResponse], error) {
	s.Log.Info("Connect: Create called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Create", "userId", userId)

	// Convert proto to domain
	protoThing := req.Msg.Thing
	if protoThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("thing is required"))
	}

	domainThing, err := ProtoThingToDomain(protoThing)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call business logic
	createdThing, err := s.BusinessService.Create(ctx, userId, *domainThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Convert back to proto
	response := &thingv1.CreateResponse{
		Thing: DomainThingToProto(createdThing),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a thing by ID
func (s *ThingConnectServer) Get(
	ctx context.Context,
	req *connect.Request[thingv1.GetRequest],
) (*connect.Response[thingv1.GetResponse], error) {
	s.Log.Info("Connect: Get called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Get", "userId", userId)

	// Parse UUID
	thingId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid thing ID format"))
	}

	// Call business logic
	thing, err := s.BusinessService.Get(ctx, thingId)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.GetResponse{
		Thing: DomainThingToProto(thing),
	}
	return connect.NewResponse(response), nil
}

// Update updates a thing
func (s *ThingConnectServer) Update(
	ctx context.Context,
	req *connect.Request[thingv1.UpdateRequest],
) (*connect.Response[thingv1.UpdateResponse], error) {
	s.Log.Info("Connect: Update called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Update", "userId", userId)

	// Parse UUID
	thingId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid thing ID format"))
	}

	// Convert proto to domain
	protoThing := req.Msg.Thing
	if protoThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("thing data is required"))
	}

	domainThing, err := ProtoThingToDomain(protoThing)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call business logic
	updatedThing, err := s.BusinessService.Update(ctx, userId, thingId, *domainThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.UpdateResponse{
		Thing: DomainThingToProto(updatedThing),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a thing
func (s *ThingConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[thingv1.DeleteRequest],
) (*connect.Response[thingv1.DeleteResponse], error) {
	s.Log.Info("Connect: Delete called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Delete", "userId", userId)

	// Parse UUID
	thingId, err := uuid.Parse(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid thing ID format"))
	}

	// Call business logic
	err = s.BusinessService.Delete(ctx, userId, thingId)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&thingv1.DeleteResponse{}), nil
}

// Search returns things based on search criteria
func (s *ThingConnectServer) Search(
	ctx context.Context,
	req *connect.Request[thingv1.SearchRequest],
) (*connect.Response[thingv1.SearchResponse], error) {
	s.Log.Info("Connect: Search called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Search", "userId", userId)

	msg := req.Msg
	params := SearchParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.Search(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.SearchResponse{
		Things: DomainThingListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Count returns the number of things
func (s *ThingConnectServer) Count(
	ctx context.Context,
	req *connect.Request[thingv1.CountRequest],
) (*connect.Response[thingv1.CountResponse], error) {
	s.Log.Info("Connect: Count called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("Count", "userId", userId)

	msg := req.Msg
	params := CountParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	count, err := s.BusinessService.Count(ctx, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.CountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}

// GeoJson returns a GeoJSON representation of things
func (s *ThingConnectServer) GeoJson(
	ctx context.Context,
	req *connect.Request[thingv1.GeoJsonRequest],
) (*connect.Response[thingv1.GeoJsonResponse], error) {
	s.Log.Info("Connect: GeoJson called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("GeoJson", "userId", userId)

	msg := req.Msg
	params := GeoJsonParams{}
	if msg.Type != 0 {
		params.Type = &msg.Type
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}
	if msg.Validated {
		params.Validated = &msg.Validated
	}

	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	result, err := s.BusinessService.GeoJson(ctx, offset, limit, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.GeoJsonResponse{
		Result: result,
	}
	return connect.NewResponse(response), nil
}

// ListByExternalId returns things filtered by external ID
func (s *ThingConnectServer) ListByExternalId(
	ctx context.Context,
	req *connect.Request[thingv1.ListByExternalIdRequest],
) (*connect.Response[thingv1.ListByExternalIdResponse], error) {
	s.Log.Info("Connect: ListByExternalId called", "externalId", req.Msg.ExternalId)

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("ListByExternalId", "userId", userId)

	msg := req.Msg
	limit := s.BusinessService.ListDefaultLimit
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.ListByExternalId(ctx, offset, limit, int(msg.ExternalId))
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	// Return NotFound if no results (matching HTTP handler behavior)
	if len(list) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no things found with this external ID"))
	}

	response := &thingv1.ListByExternalIdResponse{
		Things: DomainThingListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}
