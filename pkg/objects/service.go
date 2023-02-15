package objects

import (
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"log"
)

type Service struct {
	Log         *log.Logger
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

func (s Service) Delete(ctx echo.Context, objectId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Get(ctx echo.Context, objectId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) Update(ctx echo.Context, objectId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListByType(ctx echo.Context, typeId int32, params ListByTypeParams) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeObjectList(ctx echo.Context, params TypeObjectListParams) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeObjectCreate(ctx echo.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeObjectDelete(ctx echo.Context, typeObjectId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeObjectGet(ctx echo.Context, typeObjectId int32) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) TypeObjectUpdate(ctx echo.Context, typeObjectId int32) error {
	//TODO implement me
	panic("implement me")
}
