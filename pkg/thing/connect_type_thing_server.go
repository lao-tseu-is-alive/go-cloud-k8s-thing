// Package thing provides Connect RPC handlers for the TypeThingService.
package thing

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1/thingv1connect"
)

// TypeThingConnectServer implements the TypeThingServiceHandler interface.
// Authentication is handled by the AuthInterceptor, which injects user info into context.
type TypeThingConnectServer struct {
	BusinessService *BusinessService
	Log             *slog.Logger

	// Embed the unimplemented handler for forward compatibility
	thingv1connect.UnimplementedTypeThingServiceHandler
}

// NewTypeThingConnectServer creates a new TypeThingConnectServer.
// Note: Authentication is handled by the AuthInterceptor, not by this server.
func NewTypeThingConnectServer(business *BusinessService, log *slog.Logger) *TypeThingConnectServer {
	return &TypeThingConnectServer{
		BusinessService: business,
		Log:             log,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// mapErrorToConnect converts business errors to Connect errors
func (s *TypeThingConnectServer) mapErrorToConnect(err error) *connect.Error {
	switch {
	case errors.Is(err, ErrTypeThingNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
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
// TypeThingService RPC Methods
// =============================================================================

// List returns a list of type things
func (s *TypeThingConnectServer) List(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceListRequest],
) (*connect.Response[thingv1.TypeThingServiceListResponse], error) {
	s.Log.Info("Connect: TypeThing.List called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeThing.List", "userId", userId)

	msg := req.Msg
	params := TypeThingListParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.ExternalId != 0 {
		params.ExternalId = &msg.ExternalId
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}

	// Handle pagination
	limit := 250 // Default for TypeThing as in HTTP handler
	if msg.Limit > 0 {
		limit = int(msg.Limit)
	}
	offset := 0
	if msg.Offset > 0 {
		offset = int(msg.Offset)
	}

	list, err := s.BusinessService.ListTypeThings(ctx, offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty list instead of error
			return connect.NewResponse(&thingv1.TypeThingServiceListResponse{
				TypeThings: []*thingv1.TypeThingList{},
			}), nil
		}
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingServiceListResponse{
		TypeThings: DomainTypeThingListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new type thing
func (s *TypeThingConnectServer) Create(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceCreateRequest],
) (*connect.Response[thingv1.TypeThingServiceCreateResponse], error) {
	s.Log.Info("Connect: TypeThing.Create called")

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeThing.Create", "userId", userId, "isAdmin", isAdmin)

	protoTypeThing := req.Msg.TypeThing
	if protoTypeThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_thing is required"))
	}

	domainTypeThing := ProtoTypeThingToDomain(protoTypeThing)

	createdTypeThing, err := s.BusinessService.CreateTypeThing(ctx, userId, isAdmin, *domainTypeThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingServiceCreateResponse{
		TypeThing: DomainTypeThingToProto(createdTypeThing),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a type thing by ID
func (s *TypeThingConnectServer) Get(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceGetRequest],
) (*connect.Response[thingv1.TypeThingServiceGetResponse], error) {
	s.Log.Info("Connect: TypeThing.Get called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	_, isAdmin := GetUserFromContext(ctx)

	typeThing, err := s.BusinessService.GetTypeThing(ctx, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingServiceGetResponse{
		TypeThing: DomainTypeThingToProto(typeThing),
	}
	return connect.NewResponse(response), nil
}

// Update updates a type thing
func (s *TypeThingConnectServer) Update(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceUpdateRequest],
) (*connect.Response[thingv1.TypeThingServiceUpdateResponse], error) {
	s.Log.Info("Connect: TypeThing.Update called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeThing.Update", "userId", userId, "isAdmin", isAdmin)

	protoTypeThing := req.Msg.TypeThing
	if protoTypeThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_thing data is required"))
	}

	domainTypeThing := ProtoTypeThingToDomain(protoTypeThing)

	updatedTypeThing, err := s.BusinessService.UpdateTypeThing(ctx, userId, isAdmin, req.Msg.Id, *domainTypeThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingServiceUpdateResponse{
		TypeThing: DomainTypeThingToProto(updatedTypeThing),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a type thing
func (s *TypeThingConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceDeleteRequest],
) (*connect.Response[thingv1.TypeThingServiceDeleteResponse], error) {
	s.Log.Info("Connect: TypeThing.Delete called", "id", req.Msg.Id)

	// User info injected by AuthInterceptor
	userId, isAdmin := GetUserFromContext(ctx)
	s.Log.Info("TypeThing.Delete", "userId", userId, "isAdmin", isAdmin)

	err := s.BusinessService.DeleteTypeThing(ctx, userId, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&thingv1.TypeThingServiceDeleteResponse{}), nil
}

// Count returns the number of type things
func (s *TypeThingConnectServer) Count(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingServiceCountRequest],
) (*connect.Response[thingv1.TypeThingServiceCountResponse], error) {
	s.Log.Info("Connect: TypeThing.Count called")

	// User info injected by AuthInterceptor
	userId, _ := GetUserFromContext(ctx)
	s.Log.Info("TypeThing.Count", "userId", userId)

	msg := req.Msg
	params := TypeThingCountParams{}
	if msg.Keywords != "" {
		params.Keywords = &msg.Keywords
	}
	if msg.CreatedBy != 0 {
		params.CreatedBy = &msg.CreatedBy
	}
	if msg.Inactivated {
		params.Inactivated = &msg.Inactivated
	}

	count, err := s.BusinessService.CountTypeThings(ctx, params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingServiceCountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}
