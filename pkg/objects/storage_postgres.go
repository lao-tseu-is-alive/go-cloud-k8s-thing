package objects

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

type PGX struct {
	con *pgxpool.Pool
	log *log.Logger
}

func (P PGX) List(offset, limit int) ([]*ObjectList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Get(id int32) (*Object, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) GetMaxId() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Exist(id int32) bool {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Count() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Create(object Object) (*Object, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Update(id int32, object Object) (*Object, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Delete(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (P PGX) SearchObjectsByName(pattern string) ([]*ObjectList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) IsObjectActive(id int32) bool {
	//TODO implement me
	panic("implement me")
}

func (P PGX) CreateTypeObject(typeObject TypeObject) (*TypeObject, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) UpdateTypeObject(id int32, typeObject TypeObject) (*TypeObject, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) DeleteTypeObject(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (P PGX) ListTypeObject(offset, limit int) ([]*TypeObjectList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) GetTypeObject(id int32) (*TypeObject, error) {
	//TODO implement me
	panic("implement me")
}
