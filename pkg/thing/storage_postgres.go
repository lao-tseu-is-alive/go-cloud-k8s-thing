package thing

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

type PGX struct {
	con *pgxpool.Pool
	log golog.MyLogger
}

func (P PGX) List(offset, limit int) ([]*ThingList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Get(id int32) (*Thing, error) {
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

func (P PGX) Create(thing Thing) (*Thing, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Update(id int32, thing Thing) (*Thing, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) Delete(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (P PGX) SearchThingsByName(pattern string) ([]*ThingList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) IsThingActive(id int32) bool {
	//TODO implement me
	panic("implement me")
}

func (P PGX) CreateTypeThing(typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) UpdateTypeThing(id int32, typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) DeleteTypeThing(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (P PGX) ListTypeThing(offset, limit int) ([]*TypeThingList, error) {
	//TODO implement me
	panic("implement me")
}

func (P PGX) GetTypeThing(id int32) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}
