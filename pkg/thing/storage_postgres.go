package thing

import (
	"context"
	"errors"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

type PGX struct {
	Conn *pgxpool.Pool
	dbi  database.DB
	log  golog.MyLogger
}

// NewPgxDB will instantiate a new storage of type postgres and ensure schema exist
func NewPgxDB(db database.DB, log golog.MyLogger) (Storage, error) {
	var psql PGX
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, err
	}
	psql.Conn = pgConn
	psql.dbi = db
	psql.log = log
	var numberOfTypeThings int
	errTypeThingTable := pgConn.QueryRow(context.Background(), typeThingCount).Scan(&numberOfTypeThings)
	if errTypeThingTable != nil {
		log.Error("Unable to retrieve the number of users error: %v", err)
		return nil, err
	}

	if numberOfTypeThings > 0 {
		log.Info("'database contains %d records in «go_thing.type_thing»'", numberOfTypeThings)
	} else {
		log.Warn("«go_thing.type_thing» is empty ! it should contain at least one row")
		return nil, errors.New("problem with initial content of database «go_thing.type_thing» should not be empty ")
	}

	return &psql, err
}

// List returns the list of existing things with the given offset and limit.
func (db *PGX) List(offset, limit int) ([]*ThingList, error) {
	db.log.Debug("trace : entering List()")
	var res []*ThingList

	err := pgxscan.Select(context.Background(), db.Conn, &res, listThings, limit)
	if err != nil {
		db.log.Error("List pgxscan.Select unexpectedly failed, error : %v", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(" List returned no results ")
		return nil, errors.New("records not found")
	}

	return res, nil
}

func (db *PGX) Get(id int32) (*Thing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) GetMaxId() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) Exist(id int32) bool {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) Count() (int32, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) Create(thing Thing) (*Thing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) Update(id int32, thing Thing) (*Thing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) Delete(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) SearchThingsByName(pattern string) ([]*ThingList, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) IsThingActive(id int32) bool {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) CreateTypeThing(typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) UpdateTypeThing(id int32, typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) DeleteTypeThing(id int32) error {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) ListTypeThing(offset, limit int) ([]*TypeThingList, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) GetTypeThing(id int32) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}
