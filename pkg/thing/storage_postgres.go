package thing

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
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
func (db *PGX) List(offset, limit int, params ListParams) ([]*ThingList, error) {
	db.log.Debug("trace : entering List(params : %+v)", params)
	var (
		res []*ThingList
		err error
	)
	listThings := baseThingListQuery + listThingsConditions
	if &params.Validated != nil {
		listThings += " AND validated = coalesce(null, validated) " + thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Validated)
	} else {
		listThings += thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated)
	}
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

// ListByExternalId returns the list of existing things having given externalId with the given offset and limit.
func (db *PGX) ListByExternalId(offset, limit int, externalId int) ([]*ThingList, error) {
	db.log.Debug("trace : entering ListByExternalId(externalId : %v)", externalId)
	var res []*ThingList
	listByExternalIdThings := baseThingListQuery + listByExternalIdThingsCondition + thingListOrderBy
	err := pgxscan.Select(context.Background(), db.Conn, &res, listByExternalIdThings, limit, offset, externalId)
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

func (db *PGX) Search(offset, limit int, params SearchParams) ([]*ThingList, error) {
	db.log.Debug("trace : entering Search(params : %+v)", params)
	var (
		res []*ThingList
		err error
	)
	searchThings := baseThingListQuery + searchThingsConditions
	if &params.Validated != nil {
		searchThings += " AND validated = coalesce(null, validated) " + thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, searchThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords, &params.Validated)
	} else {
		searchThings += thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, searchThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords)
	}

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

// Get will retrieve one thing with given id
func (db *PGX) Get(id uuid.UUID) (*Thing, error) {
	db.log.Debug("trace : entering Get(%v)", id)
	res := &Thing{}
	err := pgxscan.Get(context.Background(), db.Conn, res, getThing, id)
	if err != nil {
		db.log.Error("Get(%d) pgxscan.Select unexpectedly failed, error : %v", id, err)
		return nil, err
	}
	if res == nil {
		db.log.Info(" Get(%v) returned no results ", id)
		return nil, errors.New("records not found")
	}
	return res, nil
}

// Exist returns true only if a thing with the specified id exists in store.
func (db *PGX) Exist(id uuid.UUID) bool {
	db.log.Debug("trace : entering Exist(%d)", id)
	count, err := db.dbi.GetQueryInt(existThing, id)
	if err != nil {
		db.log.Error("Exist(%v) could not be retrieved from DB. failed db.Query err: %v", id, err)
		return false
	}
	if count > 0 {
		db.log.Info(" Exist(%v) id does exist  count:%v", id, count)
		return true
	} else {
		db.log.Info(" Exist(%d) id does not exist count:%v", id, count)
		return false
	}
}

// Count returns the number of users stored in DB
func (db *PGX) Count() (int32, error) {
	db.log.Debug("trace : entering Count()")
	count, err := db.dbi.GetQueryInt(countThing)
	if err != nil {
		db.log.Error("Count() could not be retrieved from DB. failed db.Query err: %v", err)
		return 0, err
	}
	return int32(count), nil
}

// Create will store the new Thing in the database
func (db *PGX) Create(t Thing) (*Thing, error) {
	db.log.Debug("trace : entering Create(%q,%q)", t.Name, t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(createThing,
		/*	INSERT INTO go_thing.thing
			(id, type_id, name, description, comment, external_id, external_ref,
				build_at, status, contained_by, contained_by_old,validated, validated_time, validated_by,
				managed_by, _created_at, _created_by, more_data,
				text_search, position)
			VALUES ($1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14,
			$15, CURRENT_TIMESTAMP, $16, $17,
			to_tsvector('french',...)
			ST_SetSRID(ST_MakePoint($18,$19), 2056)));
		*/
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$14
		&t.ManagedBy, t.CreateBy, &t.MoreData, t.PosX, t.PosY)
	if err != nil {
		db.log.Error("Create(%q) unexpectedly failed. error : %v", t.Name, err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Create(%q) no row was created so create as failed. error : %v", t.Name, err)
		return nil, err
	}
	db.log.Info(" Create(%q) created with id : %v", t.Name, t.Id)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdThing, err := db.Get(t.Id)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error %v: thing was created, but can not be retrieved", err))
	}
	return createdThing, nil
}

func (db *PGX) Update(id uuid.UUID, t Thing) (*Thing, error) {
	db.log.Debug("trace : entering Update(%q)", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(updateTing,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Inactivated, &t.InactivatedTime, &t.InactivatedBy, &t.InactivatedReason, //$15
		t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$18
		&t.ManagedBy, &t.LastModifiedBy, &t.MoreData, t.PosX, t.PosY) //$23
	if err != nil {

		db.log.Error("Create(%q) unexpectedly failed. error : %v", t.Id, err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Create(%q) no row was created so create as failed. error : %v", t.Id, err)
		return nil, err
	}
	//db.log.Info(" Create(%q) created with id : %v", t.Name, t.Id)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedThing, err := db.Get(t.Id)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error %v: thing was updated, but can not be retrieved.", err))
	}
	return updatedThing, nil
}

// Delete the thing stored in DB with given id
func (db *PGX) Delete(id uuid.UUID, userId int32) error {
	db.log.Debug("trace : entering Delete(%d)", id)
	rowsAffected, err := db.dbi.ExecActionQuery(deleteThing, userId, id)
	if err != nil {
		msg := fmt.Sprintf("thing could not be deleted, err: %v", err)
		db.log.Error(msg)
		return errors.New(msg)
	}
	if rowsAffected < 1 {
		msg := fmt.Sprintf("thing was not deleted, err: %v", err)
		db.log.Error(msg)
		return errors.New(msg)
	}
	// if we get to here all is good
	return nil
}

func (db *PGX) SearchThingsByName(pattern string) ([]*ThingList, error) {
	//TODO implement me
	panic("implement me")
}

// IsThingActive returns true if the thing with the specified id has the inactivated attribute set to false
func (db *PGX) IsThingActive(id uuid.UUID) bool {
	db.log.Debug("trace : entering IsThingActive(%d)", id)
	count, err := db.dbi.GetQueryInt(isActiveThing, id)
	if err != nil {
		db.log.Error("IsThingActive(%d) could not be retrieved from DB. failed db.Query err: %v", id, err)
		return false
	}
	if count > 0 {
		db.log.Info(" IsThingActive(%d) is true  count:%v", id, count)
		return true
	} else {
		db.log.Info(" IsThingActive(%d) is false count:%v", id, count)
		return false
	}
}

// IsUserOwner returns true only if userId is the creator of the record (owner) of this thing in store.
func (db *PGX) IsUserOwner(id uuid.UUID, userId int32) bool {
	db.log.Debug("trace : entering IsUserOwner(%v, %d)", id, userId)
	count, err := db.dbi.GetQueryInt(existThingOwnedBy, id, userId)
	if err != nil {
		db.log.Error("IsUserOwner(%v, %d) could not be retrieved from DB. failed db.Query err: %v", id, userId, err)
		return false
	}
	if count > 0 {
		db.log.Info(" IsUserOwner(%v, %d) is true  count:%v", id, userId, count)
		return true
	} else {
		db.log.Info(" IsUserOwner(%v, %d) is false count:%v", id, userId, count)
		return false
	}
}

func (db *PGX) CreateTypeThing(typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) UpdateTypeThing(id int32, typeThing TypeThing) (*TypeThing, error) {
	//TODO implement me
	panic("implement me")
}

func (db *PGX) DeleteTypeThing(id int32, userId int32) error {
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
