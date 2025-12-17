package thing

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
)

type PGX struct {
	Conn *pgxpool.Pool
	dbi  database.DB
	log  *slog.Logger
}

// NewPgxDB will instantiate a new storage of type postgres and ensure schema exist
func NewPgxDB(ctx context.Context, db database.DB, log *slog.Logger) (Storage, error) {
	var psql PGX
	pgConn, err := db.GetPGConn()
	if err != nil {
		return nil, err
	}
	psql.Conn = pgConn
	psql.dbi = db
	psql.log = log
	var numberOfTypeThings int
	errTypeThingTable := pgConn.QueryRow(ctx, typeThingCount).Scan(&numberOfTypeThings)
	if errTypeThingTable != nil {
		log.Error("Unable to retrieve the number of typeThing", "error", err)
		return nil, errTypeThingTable
	}

	if numberOfTypeThings > 0 {
		log.Info("database contains records in go_thing.type_thing", "count", numberOfTypeThings)
	} else {
		log.Warn("go_thing.type_thing is empty - it should contain at least one row")
		return nil, fmt.Errorf("«go_thing.type_thing» contains %w should not be empty", numberOfTypeThings)
	}

	return &psql, err
}

func (db *PGX) GeoJson(ctx context.Context, offset, limit int, params GeoJsonParams) (string, error) {
	db.log.Debug("trace: entering GeoJson", "offset", offset, "limit", limit)
	if params.Type != nil {
		db.log.Info("param type", "type", *params.Type)
	}
	if params.CreatedBy != nil {
		db.log.Info("params.CreatedBy", "createdBy", *params.CreatedBy)
	}
	var (
		mayBeResultIsNull *string
		err               error
	)
	isInactive := false
	if params.Inactivated != nil {
		isInactive = *params.Inactivated
	}
	listThings := baseGeoJsonThingSearch + listThingsConditions
	if params.Validated != nil {
		db.log.Debug("params.Validated is not nil ")
		isValidated := *params.Validated
		listThings += " AND validated = coalesce($6, validated) " + geoJsonListEndOfQuery
		err = pgxscan.Select(ctx, db.Conn, &mayBeResultIsNull, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listThings += geoJsonListEndOfQuery
		err = pgxscan.Select(ctx, db.Conn, &mayBeResultIsNull, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive)
	}
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "List", err)
		return "", err
	}
	if mayBeResultIsNull == nil {
		db.log.Info("List returned no results")
		return "", pgx.ErrNoRows
	}
	return *mayBeResultIsNull, nil
}

// List returns the list of existing things with the given offset and limit.
func (db *PGX) List(ctx context.Context, offset, limit int, params ListParams) ([]*ThingList, error) {
	db.log.Debug("trace: entering List", "offset", offset, "limit", limit)
	if params.Type != nil {
		db.log.Info("param type", "type", *params.Type)
	}
	if params.CreatedBy != nil {
		db.log.Info("params.CreatedBy", "createdBy", *params.CreatedBy)
	}
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
		err = pgxscan.Select(ctx, db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive, isValidated)
	} else {
		listThings += thingListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listThings,
			limit, offset, &params.Type, &params.CreatedBy, isInactive)
	}
	if err != nil {
		db.log.Error(SelectFailedInNWithErrorE, "List", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("List returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// ListByExternalId returns the list of existing things having given externalId with the given offset and limit.
func (db *PGX) ListByExternalId(ctx context.Context, offset, limit int, externalId int) ([]*ThingList, error) {
	db.log.Debug("trace: entering ListByExternalId", "externalId", externalId)
	var res []*ThingList
	listByExternalIdThings := baseThingListQuery + listByExternalIdThingsCondition + thingListOrderBy
	err := pgxscan.Select(ctx, db.Conn, &res, listByExternalIdThings, limit, offset, externalId)
	if err != nil {
		db.log.Error("ListByExternalId failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("ListByExternalId returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

func (db *PGX) Search(ctx context.Context, offset, limit int, params SearchParams) ([]*ThingList, error) {
	db.log.Debug("trace: entering Search", "offset", offset, "limit", limit)
	var (
		res []*ThingList
		err error
	)
	searchThings := baseThingListQuery + listThingsConditions
	if params.Keywords != nil {
		searchThings += " AND text_search @@ plainto_tsquery('french', unaccent($6))"
		if params.Validated != nil {
			searchThings += " AND validated = coalesce($7, validated) " + thingListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchThings,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords, &params.Validated)
		} else {
			searchThings += thingListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchThings,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Keywords)
		}
	} else {
		if params.Validated != nil {
			searchThings += " AND validated = coalesce($6, validated) " + thingListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchThings,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated, &params.Validated)
		} else {
			searchThings += thingListOrderBy
			err = pgxscan.Select(ctx, db.Conn, &res, searchThings,
				limit, offset, &params.Type, &params.CreatedBy, &params.Inactivated)
		}
	}

	if err != nil {
		db.log.Error("Search failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("Search returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Get will retrieve the thing with given id
func (db *PGX) Get(ctx context.Context, id uuid.UUID) (*Thing, error) {
	db.log.Debug("trace: entering Get", "id", id)
	res := &Thing{}
	err := pgxscan.Get(ctx, db.Conn, res, getThing, id)
	if err != nil {
		db.log.Error("Get failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("Get returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// Exist returns true only if a thing with the specified id exists in store.
func (db *PGX) Exist(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering Exist", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, existThing, id)
	if err != nil {
		db.log.Error("Exist could not be retrieved from DB", "id", id, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("Exist: id does exist", "id", id, "count", count)
		return true
	} else {
		db.log.Info("Exist: id does not exist", "id", id, "count", count)
		return false
	}
}

// Count returns the number of thing stored in DB
func (db *PGX) Count(ctx context.Context, params CountParams) (int32, error) {
	db.log.Debug("trace : entering Count()")
	var (
		count int
		err   error
	)
	queryCount := countThing + " WHERE _deleted = false AND position IS NOT NULL "
	withoutSearchParameters := true
	if params.Keywords != nil {
		withoutSearchParameters = false
		queryCount += `AND text_search @@ plainto_tsquery('french', unaccent($1))
		AND type_id = coalesce($2, type_id)
		AND _created_by = coalesce($3, _created_by)
		AND inactivated = coalesce($4, inactivated)
`
		if params.Validated != nil {
			db.log.Debug("params.Validated is not nil ")
			isValidated := *params.Validated
			queryCount += " AND validated = coalesce($4, validated) "
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.Type, &params.CreatedBy, &params.Inactivated, isValidated)

		} else {
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.Type, &params.CreatedBy, &params.Inactivated)
		}
	}
	if withoutSearchParameters {
		queryCount += `
		AND type_id = coalesce($1, type_id)
		AND _created_by = coalesce($2, _created_by)
		AND inactivated = coalesce($3, inactivated)
`
		if params.Validated != nil {
			db.log.Debug("params.Validated is not nil ")
			isValidated := *params.Validated
			queryCount += " AND validated = coalesce($4, validated) "
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Type, &params.CreatedBy, &params.Inactivated, isValidated)

		} else {
			count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Type, &params.CreatedBy, &params.Inactivated)
		}

	}

	if err != nil {
		db.log.Error("Count failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// Create will store the new Thing in the database
func (db *PGX) Create(ctx context.Context, t Thing) (*Thing, error) {
	db.log.Debug("trace: entering Create", "name", t.Name, "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, createThing,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$14
		&t.ManagedBy, t.CreatedBy, &t.MoreData, t.PosX, t.PosY)
	if err != nil {
		db.log.Error("Create unexpectedly failed", "name", t.Name, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Create no row was created", "name", t.Name)
		return nil, err
	}
	db.log.Info("Create success", "name", t.Name, "id", t.Id)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdThing, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: thing was created, but can not be retrieved", err)
	}
	return createdThing, nil
}

// Update the thing stored in DB with given id and other information in struct
func (db *PGX) Update(ctx context.Context, id uuid.UUID, t Thing) (*Thing, error) {
	db.log.Debug("trace: entering Update", "id", t.Id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateThing,
		t.Id, t.TypeId, t.Name, &t.Description, &t.Comment, &t.ExternalId, &t.ExternalRef, //$7
		&t.BuildAt, &t.Status, &t.ContainedBy, &t.ContainedByOld, t.Inactivated, &t.InactivatedTime, &t.InactivatedBy, &t.InactivatedReason, //$15
		t.Validated, &t.ValidatedTime, &t.ValidatedBy, //$18
		&t.ManagedBy, &t.LastModifiedBy, &t.MoreData, t.PosX, t.PosY) //$23
	if err != nil {

		db.log.Error("Update unexpectedly failed", "id", t.Id, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("Update no row was updated", "id", t.Id)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedThing, err := db.Get(ctx, t.Id)
	if err != nil {
		return nil, fmt.Errorf("error %w: thing was updated, but can not be retrieved", err)
	}
	return updatedThing, nil
}

// Delete the thing stored in DB with given id
func (db *PGX) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	db.log.Debug("trace: entering Delete", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteThing, userId, id)
	if err != nil {
		db.log.Error("thing could not be deleted", "id", id, "error", err)
		return fmt.Errorf("thing could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("thing was not deleted", "id", id)
		return fmt.Errorf("thing was not marked for deletetion")
	}
	return nil
}

// IsThingActive returns true if the thing with the specified id has the inactivated attribute set to false
func (db *PGX) IsThingActive(ctx context.Context, id uuid.UUID) bool {
	db.log.Debug("trace: entering IsThingActive", "id", id)
	count, err := db.dbi.GetQueryInt(ctx, isActiveThing, id)
	if err != nil {
		db.log.Error("IsThingActive could not be retrieved from DB", "id", id, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("IsThingActive is true", "id", id, "count", count)
		return true
	} else {
		db.log.Info("IsThingActive is false", "id", id, "count", count)
		return false
	}
}

// IsUserOwner returns true only if userId is the creator of the record (owner) of this thing in store.
func (db *PGX) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) bool {
	db.log.Debug("trace: entering IsUserOwner", "id", id, "userId", userId)
	count, err := db.dbi.GetQueryInt(ctx, existThingOwnedBy, id, userId)
	if err != nil {
		db.log.Error("IsUserOwner could not be retrieved from DB", "id", id, "userId", userId, "error", err)
		return false
	}
	if count > 0 {
		db.log.Info("IsUserOwner is true", "id", id, "userId", userId, "count", count)
		return true
	} else {
		db.log.Info("IsUserOwner is false", "id", id, "userId", userId, "count", count)
		return false
	}
}

// CreateTypeThing will store the new TypeThing in the database
func (db *PGX) CreateTypeThing(ctx context.Context, tt TypeThing) (*TypeThing, error) {
	db.log.Debug("trace: entering CreateTypeThing", "name", tt.Name, "createdBy", tt.CreatedBy)
	var lastInsertId int = 0
	err := db.Conn.QueryRow(ctx, createTypeThing,
		tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, &tt.GeometryType, //$6
		&tt.ManagedBy, tt.IconPath, tt.CreatedBy, &tt.MoreDataSchema).Scan(&lastInsertId)
	if err != nil {
		db.log.Error("CreateTypeThing unexpectedly failed", "name", tt.Name, "error", err)
		return nil, err
	}
	db.log.Info("CreateTypeThing success", "name", tt.Name, "id", lastInsertId)

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	createdTypeThing, err := db.GetTypeThing(ctx, int32(lastInsertId))
	if err != nil {
		return nil, fmt.Errorf("error %w: typeThing was created, but can not be retrieved", err)
	}
	return createdTypeThing, nil
}

// UpdateTypeThing updates the TypeThing stored in DB with given id and other information in struct
func (db *PGX) UpdateTypeThing(ctx context.Context, id int32, tt TypeThing) (*TypeThing, error) {
	db.log.Debug("trace: entering UpdateTypeThing", "id", id)

	rowsAffected, err := db.dbi.ExecActionQuery(ctx, updateTypeTing,
		id, tt.Name, &tt.Description, &tt.Comment, &tt.ExternalId, &tt.TableName, //$6
		&tt.GeometryType, tt.Inactivated, &tt.InactivatedTime, &tt.InactivatedBy, &tt.InactivatedReason, //$11
		&tt.ManagedBy, tt.IconPath, &tt.LastModifiedBy, &tt.MoreDataSchema) //$14
	if err != nil {

		db.log.Error("UpdateTypeThing unexpectedly failed", "id", id, "error", err)
		return nil, err
	}
	if rowsAffected < 1 {
		db.log.Error("UpdateTypeThing no row was updated", "id", id)
		return nil, err
	}

	// if we get to here all is good, so let's retrieve a fresh copy to send it back
	updatedTypeThing, err := db.GetTypeThing(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error %w: thing was updated, but can not be retrieved", err)
	}
	return updatedTypeThing, nil
}

// DeleteTypeThing deletes the TypeThing stored in DB with given id
func (db *PGX) DeleteTypeThing(ctx context.Context, id int32, userId int32) error {
	db.log.Debug("trace: entering DeleteTypeThing", "id", id)
	rowsAffected, err := db.dbi.ExecActionQuery(ctx, deleteTypeThing, userId, id)
	if err != nil {
		db.log.Error("typething could not be deleted", "id", id, "error", err)
		return fmt.Errorf("typething could not be deleted: %w", err)
	}
	if rowsAffected < 1 {
		db.log.Error("typething was not deleted", "id", id)
		return fmt.Errorf("typething was not marked for deletion")
	}
	return nil
}

// ListTypeThing returns the list of existing TypeThing with the given offset and limit.
func (db *PGX) ListTypeThing(ctx context.Context, offset, limit int, params TypeThingListParams) ([]*TypeThingList, error) {
	db.log.Debug("trace : entering ListTypeThing")
	var (
		res []*TypeThingList
		err error
	)
	listTypeThings := typeThingListQuery
	if params.Keywords != nil {
		listTypeThings += listTypeThingsConditionsWithKeywords + typeThingListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeThings,
			limit, offset, &params.Keywords, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	} else {
		listTypeThings += listTypeThingsConditionsWithoutKeywords + typeThingListOrderBy
		err = pgxscan.Select(ctx, db.Conn, &res, listTypeThings,
			limit, offset, &params.CreatedBy, &params.ExternalId, &params.Inactivated)
	}

	if err != nil {
		db.log.Error("ListTypeThing failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("ListTypeThing returned no results")
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// GetTypeThing will retrieve the TypeThing with given id
func (db *PGX) GetTypeThing(ctx context.Context, id int32) (*TypeThing, error) {
	db.log.Debug("trace: entering GetTypeThing", "id", id)
	res := &TypeThing{}
	err := pgxscan.Get(ctx, db.Conn, res, getTypeThing, id)
	if err != nil {
		db.log.Error("GetTypeThing failed", "error", err)
		return nil, err
	}
	if res == nil {
		db.log.Info("GetTypeThing returned no results", "id", id)
		return nil, pgx.ErrNoRows
	}
	return res, nil
}

// CountTypeThing returns the number of TypeThing based on search criteria
func (db *PGX) CountTypeThing(ctx context.Context, params TypeThingCountParams) (int32, error) {
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
		count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.Keywords, &params.CreatedBy, &params.Inactivated)
	}
	if withoutSearchParameters {
		queryCount += `
		AND _created_by = coalesce($1, _created_by)
		AND inactivated = coalesce($2, inactivated)
`
		count, err = db.dbi.GetQueryInt(ctx, queryCount, &params.CreatedBy, &params.Inactivated)

	}
	if err != nil {
		db.log.Error("CountTypeThing failed", "error", err)
		return 0, err
	}
	return int32(count), nil
}

// GetTypeThingMaxId will retrieve maximum value of TypeThing id existing in store.
func (db *PGX) GetTypeThingMaxId(ctx context.Context) (int32, error) {
	db.log.Debug("trace : entering GetTypeThingMaxId")
	existingMaxId, err := db.dbi.GetQueryInt(ctx, typeThingMaxId)
	if err != nil {
		db.log.Error("GetTypeThingMaxId() failed", "error", err)
		return 0, err
	}
	return int32(existingMaxId), nil
}
