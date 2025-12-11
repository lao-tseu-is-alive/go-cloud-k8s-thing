package thing

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
)

// HTTPHandler implements the ServerInterface and handles HTTP-specific concerns
type HTTPHandler struct {
	business *BusinessService
	server   *goHttpEcho.Server
}

// NewHTTPHandler creates a new HTTP handler that wraps the business service
func NewHTTPHandler(business *BusinessService, server *goHttpEcho.Server) *HTTPHandler {
	return &HTTPHandler{
		business: business,
		server:   server,
	}
}

// mapErrorToHTTP converts business errors to appropriate HTTP responses
func (h *HTTPHandler) mapErrorToHTTP(ctx echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return ctx.JSON(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrTypeThingNotFound):
		return ctx.JSON(http.StatusNotFound, err.Error())
	case errors.Is(err, ErrAlreadyExists):
		return ctx.JSON(http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrNotOwner):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, ErrAdminRequired):
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
	case errors.Is(err, ErrInvalidInput):
		return ctx.JSON(http.StatusBadRequest, err.Error())
	case errors.Is(err, pgx.ErrNoRows):
		return ctx.JSON(http.StatusNotFound, "not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("internal error: %v", err))
	}
}

// GeoJson implements ServerInterface
func (h *HTTPHandler) GeoJson(ctx echo.Context, params GeoJsonParams) error {
	handlerName := "GeoJson"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Handle pagination
	limit := h.business.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Call business logic
	jsonResult, err := h.business.GeoJson(offset, limit, params)
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSONBlob(http.StatusOK, []byte(jsonResult))
}

// List implements ServerInterface
func (h *HTTPHandler) List(ctx echo.Context, params ListParams) error {
	handlerName := "List"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Handle pagination
	limit := h.business.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Call business logic
	list, err := h.business.List(offset, limit, params)
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, list)
}

// Create implements ServerInterface
func (h *HTTPHandler) Create(ctx echo.Context) error {
	handlerName := "Create"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Bind request body
	newThing := &Thing{}
	if err := ctx.Bind(newThing); err != nil {
		msg := fmt.Sprintf("Create has invalid format [%v]", err)
		h.business.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}

	h.business.Log.Info("Create Thing Bind ok : %+v ", newThing)

	// Call business logic
	thingCreated, err := h.business.Create(currentUserId, *newThing)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) || errors.Is(err, ErrAlreadyExists) {
			h.business.Log.Error(err.Error())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, thingCreated)
}

// Count implements ServerInterface
func (h *HTTPHandler) Count(ctx echo.Context, params CountParams) error {
	handlerName := "Count"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	numThings, err := h.business.Count(params)
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, numThings)
}

// Delete implements ServerInterface
func (h *HTTPHandler) Delete(ctx echo.Context, thingId uuid.UUID) error {
	handlerName := "Delete"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	err := h.business.Delete(currentUserId, thingId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.business.Log.Warn(err.Error())
			return ctx.JSON(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// Get implements ServerInterface
func (h *HTTPHandler) Get(ctx echo.Context, thingId uuid.UUID) error {
	handlerName := "Get"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	thing, err := h.business.Get(thingId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.business.Log.Info(err.Error())
			return ctx.JSON(http.StatusNotFound, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, thing)
}

// Update implements ServerInterface
func (h *HTTPHandler) Update(ctx echo.Context, thingId uuid.UUID) error {
	handlerName := "Update"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	h.business.Log.Info("in %s(%v) : currentUserId: %d", handlerName, thingId, currentUserId)

	// Bind request body
	updateThing := new(Thing)
	if err := ctx.Bind(updateThing); err != nil {
		msg := fmt.Sprintf("Update has invalid format error:[%v]", err)
		h.business.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}

	// Call business logic
	thingUpdated, err := h.business.Update(currentUserId, thingId, *updateThing)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			h.business.Log.Warn(err.Error())
			return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Update(%v) cannot update this id, it does not exist !", thingId))
		}
		if errors.Is(err, ErrUnauthorized) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}
		if errors.Is(err, ErrInvalidInput) {
			h.business.Log.Error(err.Error())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, thingUpdated)
}

// ListByExternalId implements ServerInterface
func (h *HTTPHandler) ListByExternalId(ctx echo.Context, externalId int32, params ListByExternalIdParams) error {
	handlerName := "ListByExternalId"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Handle pagination
	limit := h.business.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Call business logic
	list, err := h.business.ListByExternalId(offset, limit, int(externalId))
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	// ListByExternalId returns 404 when no results (different from List/Search)
	if len(list) == 0 {
		return ctx.JSON(http.StatusNotFound, list)
	}

	return ctx.JSON(http.StatusOK, list)
}

// Search implements ServerInterface
func (h *HTTPHandler) Search(ctx echo.Context, params SearchParams) error {
	handlerName := "Search"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Handle pagination
	limit := h.business.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Call business logic
	list, err := h.business.Search(offset, limit, params)
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, list)
}

// TypeThingList implements ServerInterface
func (h *HTTPHandler) TypeThingList(ctx echo.Context, params TypeThingListParams) error {
	handlerName := "TypeThingList"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Handle pagination
	limit := 250
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}

	// Call business logic
	list, err := h.business.ListTypeThings(offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ctx.JSON(http.StatusNotFound, make([]*TypeThingList, 0))
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem listing type things: %v", err))
	}

	return ctx.JSON(http.StatusOK, list)
}

// TypeThingCreate implements ServerInterface
func (h *HTTPHandler) TypeThingCreate(ctx echo.Context) error {
	handlerName := "TypeThingCreate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	isAdmin := claims.User.IsAdmin
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Bind request body
	newTypeThing := &TypeThing{
		CreatedBy: currentUserId,
	}
	if err := ctx.Bind(newTypeThing); err != nil {
		msg := fmt.Sprintf("TypeThingCreate has invalid format [%v]", err)
		h.business.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}

	// Call business logic
	typeThingCreated, err := h.business.CreateTypeThing(currentUserId, isAdmin, *newTypeThing)
	if err != nil {
		if errors.Is(err, ErrAdminRequired) {
			return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
		}
		if errors.Is(err, ErrInvalidInput) {
			h.business.Log.Error(err.Error())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusCreated, typeThingCreated)
}

// TypeThingCount implements ServerInterface
func (h *HTTPHandler) TypeThingCount(ctx echo.Context, params TypeThingCountParams) error {
	handlerName := "TypeThingCount"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	numThings, err := h.business.CountTypeThings(params)
	if err != nil {
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, numThings)
}

// TypeThingDelete implements ServerInterface
func (h *HTTPHandler) TypeThingDelete(ctx echo.Context, typeThingId int32) error {
	handlerName := "TypeThingDelete"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	isAdmin := claims.User.IsAdmin
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	err := h.business.DeleteTypeThing(currentUserId, isAdmin, typeThingId)
	if err != nil {
		if errors.Is(err, ErrAdminRequired) {
			return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
		}
		if errors.Is(err, ErrTypeThingNotFound) {
			h.business.Log.Warn(err.Error())
			return ctx.JSON(http.StatusNotFound, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// TypeThingGet implements ServerInterface
func (h *HTTPHandler) TypeThingGet(ctx echo.Context, typeThingId int32) error {
	handlerName := "TypeThingGet"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := claims.User.UserId
	isAdmin := claims.User.IsAdmin
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Call business logic
	typeThing, err := h.business.GetTypeThing(isAdmin, typeThingId)
	if err != nil {
		if errors.Is(err, ErrAdminRequired) {
			return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
		}
		if errors.Is(err, ErrTypeThingNotFound) {
			h.business.Log.Warn(err.Error())
			return ctx.JSON(http.StatusNotFound, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, typeThing)
}

// TypeThingUpdate implements ServerInterface
func (h *HTTPHandler) TypeThingUpdate(ctx echo.Context, typeThingId int32) error {
	handlerName := "TypeThingUpdate"
	goHttpEcho.TraceHttpRequest(handlerName, ctx.Request(), h.business.Log)

	// Get current user from JWT
	claims := h.server.JwtCheck.GetJwtCustomClaimsFromContext(ctx)
	currentUserId := int32(claims.User.UserId)
	isAdmin := claims.User.IsAdmin
	h.business.Log.Info("in %s : currentUserId: %d", handlerName, currentUserId)

	// Bind request body
	updateTypeThing := new(TypeThing)
	if err := ctx.Bind(updateTypeThing); err != nil {
		msg := fmt.Sprintf("TypeThingUpdate has invalid format error:[%v]", err)
		h.business.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}

	// Call business logic
	thingUpdated, err := h.business.UpdateTypeThing(currentUserId, isAdmin, typeThingId, *updateTypeThing)
	if err != nil {
		if errors.Is(err, ErrAdminRequired) {
			return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
		}
		if errors.Is(err, ErrTypeThingNotFound) {
			h.business.Log.Warn(err.Error())
			return ctx.JSON(http.StatusNotFound, fmt.Sprintf("TypeThingUpdate(%v) cannot update this id, it does not exist !", typeThingId))
		}
		if errors.Is(err, ErrInvalidInput) {
			h.business.Log.Error(err.Error())
			return ctx.JSON(http.StatusBadRequest, err.Error())
		}
		return h.mapErrorToHTTP(ctx, err)
	}

	return ctx.JSON(http.StatusOK, thingUpdated)
}
