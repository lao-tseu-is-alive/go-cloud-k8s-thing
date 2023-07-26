package thing

import (
	"fmt"
	"github.com/cristalhq/jwt/v4"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goserver"
	"net/http"
	"strings"
)

type Service struct {
	Log         golog.MyLogger
	dbConn      database.DB
	Store       Storage
	JwtSecret   []byte
	JwtDuration int
}

// List sends a list of things in the store as json based of the given filters
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
	limit := 10
	if params.Limit != nil {
		limit = int(*params.Limit)
	}
	offset := 0
	if params.Offset != nil {
		offset = int(*params.Offset)
	}
	list, err := s.Store.List(offset, limit, params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("there was a problem when calling store.List :%v", err))
	}
	return ctx.JSON(http.StatusOK, list)
}

// Create allows to insert a new thing
// curl -s -XPOST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"id": "9999971f-53d7-4eb6-8898-97f257ea5f27","type_id": 2,"name": "GilTown","description": "just a nice test","external_id": 123456789,"inactivated": false,"more_data": null,"x":2537609.0 ,"y":1152611.0   }' 'http://localhost:9090/goapi/v1/thing'
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
		CreateBy: int32(currentUserId),
	}
	if err := ctx.Bind(newThing); err != nil {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Create has invalid format [%v]", err))
	}
	if len(strings.Trim(newThing.Name, " ")) < 1 {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprint("Create name cannot be empty or spaces only"))
	}
	if len(newThing.Name) < 5 {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Create name minLength is 5 not (%d)", len(newThing.Name)))
	}
	s.Log.Info("# Create() newThing : %#v\n", newThing)
	thingCreated, err := s.Store.Create(*newThing)
	if err != nil {
		msg := fmt.Sprintf("Create had an error saving thing:%#v, err:%#v", *newThing, err)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusBadRequest, msg)
	}
	s.Log.Info("# Create() Thing %#v\n", thingCreated)
	return ctx.JSON(http.StatusCreated, thingCreated)
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
	if s.Store.Exist(thingId) == false {
		msg := fmt.Sprintf("Delete(%v) cannot delete this id, it does not exist !", thingId)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	} else {
		err := s.Store.Delete(thingId, currentUserId)
		if err != nil {
			msg := fmt.Sprintf("Delete(%v) got an error: %#v ", thingId, err)
			s.Log.Info(msg)
			return echo.NewHTTPError(http.StatusInternalServerError, msg)
		}
		return ctx.NoContent(http.StatusNoContent)
	}
}

// Get will retrieve the Thing in the store and return it
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
	currentUserId := claims.Id
	if !s.Store.IsUserOwner(thingId, currentUserId) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user is not owner of this thing")
	}
	/* TODO implement ACL & RBAC handling
	if !s.Store.IsUserAllowedToDelete(currentUserId, typeThing) {
		return echo.NewHTTPError(http.StatusUnauthorized, "current user has no create role privilege")
	}
	*/
	if s.Store.Exist(thingId) == false {
		msg := fmt.Sprintf("Get(%v) cannot delete this id, it does not exist !", thingId)
		s.Log.Info(msg)
		return ctx.JSON(http.StatusNotFound, msg)
	}

	thing, err := s.Store.Get(thingId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("problem retrieving thing :%v", err))
	}
	return ctx.JSON(http.StatusOK, thing)
}

func (s Service) Update(ctx echo.Context, thingId uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListByExternalId(ctx echo.Context, externalId int32, params ListByExternalIdParams) error {
	//TODO implement me
	panic("implement me")
}
func (s Service) TypeThingList(ctx echo.Context, params TypeThingListParams) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingCreate(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingDelete(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingGet(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeThingUpdate(ctx echo.Context, typeThingId int32) error {
	//TODO implement me
	panic("implement me")
}
