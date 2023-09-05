package thing

import (
	"context"
	"errors"
	"fmt"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
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
	isInactive := false
	if params.Inactivated != nil {
		isInactive = *params.Inactivated
	}
	listThings := baseThingListQuery + listThingsConditions
	if params.Validated != nil {
		db.log.Debug("params.Validated is not nil ")
		isValidated := *params.Validated
		listThings += " AND validated = coalesce($6, validated) " + thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listThings += thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive)
	}
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "List", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "List")
		return nil, pgx.ErrNoRows
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
		db.log.Error(SelectFailedInNWithErrorE, "ListByExternalId", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "ListByExternalId")
		return nil, pgx.ErrNoRows
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
	if params.Validated != nil {
		searchThings += " AND validated = coalesce($7, validated) " + thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, searchThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords, &params.Validated)
	} else {
		searchThings += thingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, searchThings,
			limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords)
	}

	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "Search", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "Search")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Get will retrieve the thing with given id
func (db *PGX) Get(id uuid.UUID) (*Thing, error) {
	db.log.Debug("trace : entering Get(%v)", id)
	res := &Thing{}
	err := pgxscan.Get(context.Background(), db.Conn, res, getThing, id)
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "Get", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "Get")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Exist returns true only if a thing with the specified id exists in store.
func (db *PGX) Exist(id uuid.UUID) bool {
	db.log.Debug("trace : entering Exist(%v)", id)
	count, err := db.dbi.GetQueryInt(existThing, id)
	if err != nil {
		db.log.Error("Exist(%v) could not be retrieved from DB. failed db.Query err: %v", id, err)
		return false
	}
	if count > 0 {
		db.log.Info(" Exist(%v) id does exist  count:%v", id, count)
		return true
	} else {
		db.log.Info(" Exist(%v) id does not exist count:%v", id, count)
		return false
	}
}

// Count returns the number of thing stored in DB
func (db *PGX) Count(params CountParams) (int32, error) {
	db.log.Debug("trace : entering Count()")
	var (
		count int
		err   error
	)
	queryCount := countThing + " WHERE 1 = 1 "
	withoutSearchParameters := true
	if params.Keywords != nil {
		withoutSearchParameters = false
		queryCount += `AND text_search @@ plainto_tsquery('french', unaccent($1))
		AND type_id = coalesce($2, type_id)
		AND _created_by = coalesce($3, _created_by)
		AND inactivated = coalesce($4, inactivated)
`
		count, err = db.dbi.GetQueryInt(queryCount, &params.Keywords, &params.Type, &params.CreatedBy, &params.Inactivated)
	}
	if withoutSearchParameters {
		queryCount += `
		AND type_id = coalesce($1, type_id)
		AND _created_by = coalesce($2, _created_by)
		AND inactivated = coalesce($3, inactivated)
`
		count, err = db.dbi.GetQueryInt(queryCount, &params.Type, &params.CreatedBy, &params.Inactivated)

	}
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
		&t.ManagedBy, t.CreatedBy, &t.MoreData, t.PosX, t.PosY)
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

// Update the thing stored in DB with given id and other information in struct
func (db *PGX) Update(id uuid.UUID, t Thing) (*Thing, error) {
	db.log.Debug("trace : entering Update(%q)", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(updateThing,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Inactivated, &t.InactivatedTime, &t.InactivatedBy, &t.InactivatedReason, //$15
		t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$18
		&t.ManagedBy, &t.LastModifiedBy, &t.MoreData, t.PosX, t.PosY) //$23
	if err != nil {

		db.log.Error("Update(%q) unexpectedly failed. error : %v", t.Id, err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Update(%q) no row was created so create as failed. error : %v", t.Id, err)
		return nil, err
	}

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

// CreateTypeThing will store the new TypeThing in the database
func (db *PGX) CreateTypeThing(tt TypeThing) (*TypeThing, error) {
	db.log.Debug("trace : entering CreateTypeThing(%q, userId)", tt.Name, tt.CreatedBy)
	var lastInsertId int = 0
	err := db.Conn.QueryRow(context.Background(), createTypeThing,
		/*	INSERT INTO go_thing.type_thing
			    (name, description, comment, external_id, table_name, geometry_type,
			     managed_by, _created_at, _created_by, more_data_schema, text_search)
			VALUES ($1, $2, $3, $4, $5, $6,
			        $7, CURRENT_TIMESTAMP, $8, $9,
			        to_tsvector('french', unaccent($1) ||
			                              ' ' || coalesce(unaccent($2), ' ') ||
			                              ' ' || coalesce(unaccent($3), ' ') ));
		*/
		tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, &tt.GeometryType, //$6
		&tt.ManagedBy, tt.CreatedBy, &tt.MoreDataSchema).Scan(&lastInsertId)
	if err != nil {
		db.log.Error("CreateTypeThing(%q) unexpectedly failed. error : %v", tt.Name, err)
		return nil, err
	}
	db.log.Info(" CreateTypeThing(%q) created with id : %v", tt.Name, &lastInsertId)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdTypeThing, err := db.GetTypeThing(int32(lastInsertId))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error %v: typeThing was created, but can not be retrieved", err))
	}
	return createdTypeThing, nil
}

// UpdateTypeThing updates the TypeThing stored in DB with given id and other information in struct
func (db *PGX) UpdateTypeThing(id int32, tt TypeThing) (*TypeThing, error) {
	db.log.Debug("trace : entering UpdateTypeThing(%d)", id)

	rowsAffected, err := db.dbi.ExecActionQuery(updateTypeTing,
		/*		UPDATE go_thing.type_thing
				SET
				    name               = $2,
				    description        = $3,
				    comment            = $4,
				    external_id        = $5,
				    table_name         = $6,
				    geometry_type      = $7,
				    inactivated        = $8,
				    inactivated_time   = $9,
				    inactivated_by     = $10,
				    inactivated_reason = $11,
				    managed_by         = $12,
				    _last_modified_at  = CURRENT_TIMESTAMP,
				    _last_modified_by  = $13,
				    more_data_schema   = $14,
				    text_search = to_tsvector('french', unaccent($2) ||
				                             ' ' || coalesce(unaccent($3), ' ') ||
				                             ' ' || coalesce(unaccent($4), ' ') )
				WHERE id = $1;
		*/
		id, tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, //$6
		&tt.GeometryType, tt.Inactivated, &tt.InactivatedTime, &tt.InactivatedBy, &tt.InactivatedReason, //$11
		&tt.ManagedBy, &tt.LastModifiedBy, &tt.MoreDataSchema) //$14
	if err != nil {

		db.log.Error("UpdateTypeThing(%q) unexpectedly failed. error : %v", id, err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("UpdateTypeThing(%q) no row was created so create as failed. error : %v", id, err)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedTypeThing, err := db.GetTypeThing(id)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error %v: thing was updated, but can not be retrieved.", err))
	}
	return updatedTypeThing, nil
}

// DeleteTypeThing deletes the TypeThing stored in DB with given id
func (db *PGX) DeleteTypeThing(id int32, userId int32) error {
	db.log.Debug("trace : entering DeleteTypeThing(%d)", id)
	rowsAffected, err := db.dbi.ExecActionQuery(deleteTypeThing, userId, id)
	if err != nil {
		msg := fmt.Sprintf("typething could not be deleted, err: %v", err)
		db.log.Error(msg)
		return errors.New(msg)
	}
	if rowsAffected < 1 {
		msg := fmt.Sprintf("typething was not deleted, err: %v", err)
		db.log.Error(msg)
		return errors.New(msg)
	}
	// if we get to here all is good
	return nil
}

// ListTypeThing returns the list of existing TypeThing with the given offset and limit.
func (db *PGX) ListTypeThing(offset, limit int, params TypeThingListParams) ([]*TypeThingList, error) {
	db.log.Debug("trace : entering ListTypeThing")
	var (
		res []*TypeThingList
		err error
	)
	listTypeThings := typeThingListQuery
	if params.Keywords != nil {
		listTypeThings += listTypeThingsConditionsWithKeywords + typeThingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listTypeThings,
			limit, offset, &params.Keywords, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	} else {
		listTypeThings += listTypeThingsConditionsWithoutKeywords + typeThingListOrderBy
		err = pgxscan.Select(context.Background(), db.Conn, &res, listTypeThings,
			limit, offset, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	}

	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "ListTypeThing", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "ListTypeThing")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// GetTypeThing will retrieve the TypeThing with given id
func (db *PGX) GetTypeThing(id int32) (*TypeThing, error) {
	db.log.Debug("trace : entering GetTypeThing(%v)", id)
	res := &TypeThing{}
	err := pgxscan.Get(context.Background(), db.Conn, res, getTypeThing, id)
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "GetTypeThing", err)
		return nil, err
	}
	if res == nil {
		db.log.Info(FunctionNReturnedNoResults, "GetTypeThing", id)
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// CountTypeThing returns the number of TypeThing based on search criteria
func (db *PGX) CountTypeThing(params TypeThingCountParams) (int32, error) {
	db.log.Debug("trace : entering CountTypeThing()")
	var (
		count int
		err   error
	)
	queryCount := countTypeThing + " WHERE 1 = 1 "
	withoutSearchParameters := true
	if params.Keywords != nil {
		withoutSearchParameters = false
		queryCount += `AND text_search @@ plainto_tsquery('french', unaccent($1))
		AND _created_by = coalesce($2, _created_by)
		AND inactivated = coalesce($3, inactivated)
`
		count, err = db.dbi.GetQueryInt(queryCount, &params.Keywords, &params.CreatedBy, &params.Inactivated)
	}
	if withoutSearchParameters {
		queryCount += `
		AND _created_by = coalesce($1, _created_by)
		AND inactivated = coalesce($2, inactivated)
`
		count, err = db.dbi.GetQueryInt(queryCount, &params.CreatedBy, &params.Inactivated)

	}
	if err != nil {
		db.log.Error("Count() could not be retrieved from DB. failed db.Query err: %v", err)
		return 0, err
	}
	return int32(count), nil
}

// GetTypeThingMaxId will retrieve maximum value of TypeThing id existing in store.
func (db *PGX) GetTypeThingMaxId(id int32) (int32, error) {
	db.log.Debug("trace : entering GetTypeThingMaxId(%v)")
	existingMaxId, err := db.dbi.GetQueryInt(typeThingMaxId)
	if err != nil {
		db.log.Error("GetTypeThingMaxId() failed, error : %v", err)
		return 0, err
	}
	return int32(existingMaxId), nil
}
