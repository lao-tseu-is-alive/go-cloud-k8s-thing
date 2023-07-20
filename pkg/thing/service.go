package thing

import (
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

type Service struct {
	Log         golog.MyLogger
	dbConn      database.DB
	Store       Storage
	JwtSecret   []byte
	JwtDuration int
}

func (s Service) List(ctx echo.Context, params ListParams) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Create(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Delete(ctx echo.Context, thingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Get(ctx echo.Context, thingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Update(ctx echo.Context, thingId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListByType(ctx echo.Context, typeId int32, params ListByTypeParams) error {
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
