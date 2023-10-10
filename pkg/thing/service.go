package thing

import (
	"errors"
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goserver"
	"net/http"
	"strings"
)

type Permission int8 // enum
const (
	R Permission = iota // Read implies List (SELECT in DB, or GET in API)
	W                   // implies INSERT,UPDATE, DELETE
	M                   // Update or Put only
	D                   // Delete only
	C                   // Create only (Insert, Post)
	P                   // change Permissions of one thing
	O                   // change Owner of one Thing
	A                   // Audit log of changes of one thing and read only special _fields like _created_by
)

func (s Permission) String() string {
	switch s {
	case R:
		return "R"
	case W:
		return "W"
	case M:
		return "M"
	case D:
		return "D"
	case C:
		return "C"
	case P:
		return "P"
	case O:
		return "O"
	case A:
		return "A"
	}
	return "ErrorPermissionUnknown"
}

type Service struct {
	Log              golog.MyLogger
	DbConn           database.DB
	Store            Storage
	JwtSecret        []byte
	JwtDuration      int
	ListDefaultLimit int
}

// List sends a list of things in the store based on the given parameters filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing?limit=3&type=112' |jq
func (s Service) List(ctx echo.Context, params ListParams) error {
	s.Log.Info("trace: entering Thing List() params:%+v", params)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.List(offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.List :%v", err))
		} else {
			list = make([]*ThingList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Create allows to insert a new thing
// curl -s -XPOST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": NULL,"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/thing'
func (s Service) Create(ctx echo.Context) error {
	s.Log.Debug("trace: entering Create()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT ADMIN RETURN 401 Unauthorized
	currentUserId := claims.Id
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToCreate(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	newThing := &Thing{
		CreatedBy: int32(currentUserId),
	}
	if err := ctx.Bind(newThing); err != nil {
		msg := fmt.Sprintf("Create has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("Create Thing Bind ok : %+v ", newThing)
	if len(strings.Trim(newThing.Name, " ")) < 1 {

		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newThing.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN, "name", MinNameLength)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if s.Store.Exist(newThing.Id) {
		msg := fmt.Sprintf("This id (%v) already exist !", newThing.Id)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	thingCreated, err := s.Store.Create(*newThing)
	if err != nil {
		msg := fmt.Sprintf("Create had an error saving thing:%#v, err:%#v", *newThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# Create() success Thing %#v\n", thingCreated)
	return ctx.JSON(http.StatusCreated, thingCreated)
}

// Count returns the number of things found after filtering data with any given CountParams
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing/count' |jq
func (s Service) Count(ctx echo.Context, params CountParams) error {
	s.Log.Info("trace: entering Count()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	// currentUserId := claims.Id
	numThings, err := s.Store.Count(params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting things :%v", err))
	}
	return ctx.JSON(http.StatusOK, numThings)
}

// Delete will remove the given thingId entry from the store, and if not present will return 400 Bad Request
// curl -v -XDELETE -H "Content-Type: application/json" -H "Authorization: Bearer $token" 'http://localhost:8888/api/users/3' ->  204 No Content if present and delete it
// curl -v -XDELETE -H "Content-Type: application/json"  -H "Authorization: Bearer $token" 'http://localhost:8888/users/93333' -> 400 Bad Request
func (s Service) Delete(ctx echo.Context, thingId uuid.UUID) error {
	s.Log.Info("trace: entering Delete(%v)", thingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	if s.Store.Exist(thingId) == false {
		msg := fmt.Sprintf("Delete(%v) cannot delete this id, it does not exist !", thingId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	currentUserId := claims.Id
	if !s.Store.IsUserOwner(thingId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this thing")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToDelete(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	err = s.Store.Delete(thingId, currentUserId)
	if err != nil {
		msg := fmt.Sprintf("Delete(%v) got an error: %#v ", thingId, err)
		s.Log.Error(msg)
		return echo.NewHTTPError(http.StatusInternalServerError, msg)
	}
	return ctx.NoContent(http.StatusNoContent)

}

// Get will retrieve the Thing with the given id in the store and return it
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing/9999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Get(ctx echo.Context, thingId uuid.UUID) error {
	s.Log.Info("trace: entering Get(%v)", thingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	// currentUserId := claims.Id
	if s.Store.Exist(thingId) == false {
		msg := fmt.Sprintf("Get(%v) cannot get this id, it does not exist !", thingId)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToDelete(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/

	thing, err := s.Store.Get(thingId)
	if err != nil {
		if err != pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving thing :%v", err))
		} else {
			msg := fmt.Sprintf("Get(%v) no rows found in db", thingId)
			s.Log.Info(msg)
			return ctx.JSON(http.StatusNotFound, msg)
		}
	}
	return ctx.JSON(http.StatusOK, thing)
}

// Update will change the attributes values for the thing identified by the given thingId
// curl -s -XPUT -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "3999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 3,"name": "Gil-Parcelle","description": "just a nice parcelle test by GIL","external_id": 345678912,"inactivated": false,"managed_by": 999, "more_data": {"info_value": 3230 },"pos_x":2537603.0 ,"pos_y":1152613.0   }' 'http://localhost:9090/goapi/v1/thing/3999971f-53d7-4eb6-8898-97f257ea5f27' |jq
func (s Service) Update(ctx echo.Context, thingId uuid.UUID) error {
	s.Log.Debug("trace: entering Update(id=%v)", thingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	currentUserId := claims.Id
	if s.Store.Exist(thingId) == false {
		msg := fmt.Sprintf("Update(%v) cannot update this id, it does not exist !", thingId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	if !s.Store.IsUserOwner(thingId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this thing")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToCreate(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/

	updateThing := new(Thing)
	if err := ctx.Bind(updateThing); err != nil {
		msg := fmt.Sprintf("Update has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(updateThing.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(updateThing.Name) < MinNameLength {

		msg := fmt.Sprintf(FieldMinLengthIsN+FoundNum, "name", MinNameLength, len(updateThing.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	updateThing.LastModifiedBy = &currentUserId
	//TODO handle update of validated field correctly by adding validated time & user
	// handle update of managed_by field correctly by checking if user is a valid active one
	thingUpdated, err := s.Store.Update(thingId, *updateThing)
	if err != nil {
		msg := fmt.Sprintf("Update had an error saving thing:%#v, err:%#v", *updateThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# Update success Thing %#v\n", thingUpdated)
	return ctx.JSON(http.StatusOK, thingUpdated)
}

// ListByExternalId sends a list of things in the store as json based of the given filters
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing/by-external-id/345678912?limit=3&ofset=0' |jq
func (s Service) ListByExternalId(ctx echo.Context, externalId int32, params ListByExternalIdParams) error {
	s.Log.Info("trace: entering Thing ListByExternalId() externalId:%+v", externalId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.ListByExternalId(offset, limit, int(externalId))
	if err != nil {
		if err != pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.ListByExternalId :%v", err))
		} else {
			list = make([]*ThingList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// Search returns a list of things in the store as json based of the given search criteria
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing/search?limit=3&ofset=0' |jq
// curl -s -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" 'http://localhost:9090/goapi/v1/thing/search?limit=3&type=112' |jq
func (s Service) Search(ctx echo.Context, params SearchParams) error {
	s.Log.Info("trace: entering Thing Search() params:%+v", params)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.Search(offset, limit, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			list = make([]*ThingList, 0)
			return ctx.JSON(http.StatusOK, list)
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.Search :%v", err))
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeThingList sends a list of TypeThing based on the given TypeThingListParams parameters filters
func (s Service) TypeThingList(ctx echo.Context, params TypeThingListParams) error {
	s.Log.Info("trace: entering TypeThingList() params:%+v", params)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	limit := s.ListDefaultLimit
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.ListTypeThing(offset, limit, params)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.ListTypeThing :%v", err))
		} else {
			list = make([]*TypeThingList, 0)
			return ctx.JSON(http.StatusNotFound, list)
		}
	}
	return ctx.JSON(http.StatusOK, list)
}

// TypeThingCreate will insert a new TypeThing in the store
func (s Service) TypeThingCreate(ctx echo.Context) error {
	s.Log.Debug("trace: entering TypeThingCreate()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	currentUserId := claims.Id
	if !claims.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
	}
	newTypeThing := &TypeThing{
		Comment:           nil,
		CreatedAt:         nil,
		CreatedBy:         int32(currentUserId),
		Deleted:           false,
		DeletedAt:         nil,
		DeletedBy:         nil,
		Description:       nil,
		ExternalId:        nil,
		GeometryType:      nil,
		Id:                0,
		Inactivated:       false,
		InactivatedBy:     nil,
		InactivatedReason: nil,
		InactivatedTime:   nil,
		LastModifiedAt:    nil,
		LastModifiedBy:    nil,
		ManagedBy:         nil,
		MoreDataSchema:    nil,
		Name:              "",
		TableName:         nil,
	}
	if err := ctx.Bind(newTypeThing); err != nil {
		msg := fmt.Sprintf("TypeThingCreate has invalid format [%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(newTypeThing.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(newTypeThing.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(newTypeThing.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	//s.Log.Info("# Create() before Store.TypeThingCreate newThing : %#v\n", newThing)
	typeThingCreated, err := s.Store.CreateTypeThing(*newTypeThing)
	if err != nil {
		msg := fmt.Sprintf("TypeThingCreate had an error saving thing:%#v, err:%#v", *newTypeThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# TypeThingCreate() success TypeThing %#v\n", typeThingCreated)
	return ctx.JSON(http.StatusCreated, typeThingCreated)
}

func (s Service) TypeThingCount(ctx echo.Context, params TypeThingCountParams) error {
	s.Log.Info("trace: entering TypeThingCount()")
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT OWNER OF RECORD RETURN 401 Unauthorized
	// currentUserId := claims.Id
	numThings, err := s.Store.CountTypeThing(params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem counting things :%v", err))
	}
	return ctx.JSON(http.StatusOK, numThings)
}

// TypeThingDelete will remove the given TypeThing entry from the store, and if not present will return 400 Bad Request
func (s Service) TypeThingDelete(ctx echo.Context, typeThingId int32) error {
	s.Log.Debug("trace: entering TypeThingUpdate(id=%v)", typeThingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	currentUserId := claims.Id
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
	}
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		msg := fmt.Sprintf("TypeThingDelete(%v) cannot delete this id, it does not exist !", typeThingId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	} else {
		err := s.Store.DeleteTypeThing(typeThingId, currentUserId)
		if err != nil {
			msg := fmt.Sprintf("TypeThingDelete(%v) got an error: %#v ", typeThingId, err)
			s.Log.Error(msg)
			return echo.NewHTTPError(http.StatusInternalServerError, msg)
		}
		return ctx.NoContent(http.StatusNoContent)
	}
}

// TypeThingGet will retrieve the Thing with the given id in the store and return it
func (s Service) TypeThingGet(ctx echo.Context, typeThingId int32) error {
	s.Log.Info("trace: entering TypeThingGet(%v)", typeThingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// currentUserId := claims.Id
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	if !claims.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
	}
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		msg := fmt.Sprintf("TypeThingGet(%v) cannot retrieve this id, it does not exist !", typeThingId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}

	typeThing, err := s.Store.GetTypeThing(typeThingId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving TypeThing :%v", err))
	}
	return ctx.JSON(http.StatusOK, typeThing)
}

func (s Service) TypeThingUpdate(ctx echo.Context, typeThingId int32) error {
	s.Log.Debug("trace: entering TypeThingUpdate(id=%v)", typeThingId)
	// get the current user from JWT TOKEN
	u := ctx.Get("jwtdata").(*jwt.Token)
	claims := goserver.JwtCustomClaims{}
	err := u.DecodeClaims(&claims)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	// IF USER IS NOT ADMIN  RETURN 401 Unauthorized
	currentUserId := claims.Id
	if !claims.IsAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, OnlyAdminCanManageTypeThings)
	}
	typeThingCount, err := s.DbConn.GetQueryInt(existTypeThing, typeThingId)
	if err != nil || typeThingCount < 1 {
		msg := fmt.Sprintf("TypeThingUpdate(%v) cannot update this id, it does not exist !", typeThingId)
		s.Log.Warn(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}
	uTypeThing := new(TypeThing)
	if err := ctx.Bind(uTypeThing); err != nil {
		msg := fmt.Sprintf("TypeThingUpdate has invalid format error:[%v]", err)
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(strings.Trim(uTypeThing.Name, " ")) < 1 {
		msg := fmt.Sprintf(FieldCannotBeEmpty, "name")
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	if len(uTypeThing.Name) < MinNameLength {
		msg := fmt.Sprintf(FieldMinLengthIsN+", found %d", "name", MinNameLength, len(uTypeThing.Name))
		s.Log.Error(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	uTypeThing.LastModifiedBy = &currentUserId
	thingUpdated, err := s.Store.UpdateTypeThing(typeThingId, *uTypeThing)
	if err != nil {
		msg := fmt.Sprintf("TypeThingUpdate had an error saving typeThing:%#v, err:%#v", *uTypeThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# TypeThingUpdate success updating TypeThing %#+v\n", thingUpdated)
	return ctx.JSON(http.StatusOK, thingUpdated)
}
