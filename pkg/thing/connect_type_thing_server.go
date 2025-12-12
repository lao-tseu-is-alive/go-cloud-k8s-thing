// Package thing provides Connect RPC handlers for the TypeThingService.
package thing

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1/thingv1connect"
)

// TypeThingConnectServer implements the TypeThingServiceHandler interface
type TypeThingConnectServer struct {
	BusinessService *BusinessService
	Log             golog.MyLogger
	JwtCheck        goHttpEcho.JwtChecker

	// Embed the unimplemented handler for forward compatibility
	thingv1connect.UnimplementedTypeThingServiceHandler
}

// NewTypeThingConnectServer creates a new TypeThingConnectServer
func NewTypeThingConnectServer(business *BusinessService, log golog.MyLogger, jwtCheck goHttpEcho.JwtChecker) *TypeThingConnectServer {
	return &TypeThingConnectServer{
		BusinessService: business,
		Log:             log,
		JwtCheck:        jwtCheck,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// getUserFromContext extracts user info from the Authorization header
func (s *TypeThingConnectServer) getUserFromContext(ctx context.Context, header http.Header) (userId int32, isAdmin bool, err error) {
	auth := header.Get("Authorization")
	if auth == "" {
		return 0, false, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authorization header"))
	}

	// Extract Bearer token
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == auth {
		return 0, false, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid authorization format"))
	}

	// Validate token and extract claims
	claims, err := s.JwtCheck.ParseToken(token)
	if err != nil {
		s.Log.Warn("invalid JWT token: %v", err)
		return 0, false, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
	}

	return int32(claims.User.UserId), claims.User.IsAdmin, nil
}

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
		s.Log.Error("internal error: %v", err)
		return connect.NewError(connect.CodeInternal, errors.New("internal error"))
	}
}

// =============================================================================
// TypeThingService RPC Methods
// =============================================================================

// List returns a list of type things
func (s *TypeThingConnectServer) List(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingListRequest],
) (*connect.Response[thingv1.TypeThingListResponse], error) {
	s.Log.Info("Connect: TypeThing.List called")

	userId, _, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	s.Log.Info("TypeThing.List: userId=%d", userId)

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

	list, err := s.BusinessService.ListTypeThings(offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return empty list instead of error
			return connect.NewResponse(&thingv1.TypeThingListResponse{
				TypeThings: []*thingv1.TypeThingList{},
			}), nil
		}
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingListResponse{
		TypeThings: DomainTypeThingListSliceToProto(list),
	}
	return connect.NewResponse(response), nil
}

// Create creates a new type thing
func (s *TypeThingConnectServer) Create(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingCreateRequest],
) (*connect.Response[thingv1.TypeThingCreateResponse], error) {
	s.Log.Info("Connect: TypeThing.Create called")

	userId, isAdmin, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	s.Log.Info("TypeThing.Create: userId=%d, isAdmin=%v", userId, isAdmin)

	protoTypeThing := req.Msg.TypeThing
	if protoTypeThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_thing is required"))
	}

	domainTypeThing := ProtoTypeThingToDomain(protoTypeThing)

	createdTypeThing, err := s.BusinessService.CreateTypeThing(userId, isAdmin, *domainTypeThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingCreateResponse{
		TypeThing: DomainTypeThingToProto(createdTypeThing),
	}
	return connect.NewResponse(response), nil
}

// Get retrieves a type thing by ID
func (s *TypeThingConnectServer) Get(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingGetRequest],
) (*connect.Response[thingv1.TypeThingGetResponse], error) {
	s.Log.Info("Connect: TypeThing.Get called for id=%d", req.Msg.Id)

	_, isAdmin, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}

	typeThing, err := s.BusinessService.GetTypeThing(isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingGetResponse{
		TypeThing: DomainTypeThingToProto(typeThing),
	}
	return connect.NewResponse(response), nil
}

// Update updates a type thing
func (s *TypeThingConnectServer) Update(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingUpdateRequest],
) (*connect.Response[thingv1.TypeThingUpdateResponse], error) {
	s.Log.Info("Connect: TypeThing.Update called for id=%d", req.Msg.Id)

	userId, isAdmin, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	s.Log.Info("TypeThing.Update: userId=%d, isAdmin=%v", userId, isAdmin)

	protoTypeThing := req.Msg.TypeThing
	if protoTypeThing == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("type_thing data is required"))
	}

	domainTypeThing := ProtoTypeThingToDomain(protoTypeThing)

	updatedTypeThing, err := s.BusinessService.UpdateTypeThing(userId, isAdmin, req.Msg.Id, *domainTypeThing)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingUpdateResponse{
		TypeThing: DomainTypeThingToProto(updatedTypeThing),
	}
	return connect.NewResponse(response), nil
}

// Delete deletes a type thing
func (s *TypeThingConnectServer) Delete(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingDeleteRequest],
) (*connect.Response[thingv1.TypeThingDeleteResponse], error) {
	s.Log.Info("Connect: TypeThing.Delete called for id=%d", req.Msg.Id)

	userId, isAdmin, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	s.Log.Info("TypeThing.Delete: userId=%d, isAdmin=%v", userId, isAdmin)

	err = s.BusinessService.DeleteTypeThing(userId, isAdmin, req.Msg.Id)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	return connect.NewResponse(&thingv1.TypeThingDeleteResponse{}), nil
}

// Count returns the number of type things
func (s *TypeThingConnectServer) Count(
	ctx context.Context,
	req *connect.Request[thingv1.TypeThingCountRequest],
) (*connect.Response[thingv1.TypeThingCountResponse], error) {
	s.Log.Info("Connect: TypeThing.Count called")

	userId, _, err := s.getUserFromContext(ctx, req.Header())
	if err != nil {
		return nil, err
	}
	s.Log.Info("TypeThing.Count: userId=%d", userId)

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

	count, err := s.BusinessService.CountTypeThings(params)
	if err != nil {
		return nil, s.mapErrorToConnect(err)
	}

	response := &thingv1.TypeThingCountResponse{
		Count: count,
	}
	return connect.NewResponse(response), nil
}
