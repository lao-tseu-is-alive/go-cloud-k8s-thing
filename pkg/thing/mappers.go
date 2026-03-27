package thing

import (
	"time"

	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Helper Functions
// =============================================================================

// timeToTimestamp converts a *time.Time to *timestamppb.Timestamp
func timeToTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

// timestampToTime converts a *timestamppb.Timestamp to *time.Time
func timestampToTime(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

// stringPtr returns a pointer to the string, or nil if empty
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefString safely dereferences a string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// int32Ptr returns a pointer to the int32, or nil if zero
func int32Ptr(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

// derefInt32 safely dereferences an int32 pointer, returning 0 if nil
func derefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// boolPtr returns a pointer to the bool
func boolPtr(b bool) *bool {
	return &b
}

// derefBool safely dereferences a bool pointer, returning false if nil
func derefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// mapToStruct converts a map[string]interface{} to *structpb.Struct
func mapToStruct(m *map[string]interface{}) *structpb.Struct {
	if m == nil {
		return nil
	}
	s, err := structpb.NewStruct(*m)
	if err != nil {
		return nil
	}
	return s
}

// structToMap converts a *structpb.Struct to *map[string]interface{}
func structToMap(s *structpb.Struct) *map[string]interface{} {
	if s == nil {
		return nil
	}
	m := s.AsMap()
	return &m
}

// statusToString converts a *string to string
func statusToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringToStatus converts a string to *string
func stringToStatus(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// =============================================================================
// Thing Mappers
// =============================================================================

// DomainThingToProto converts a domain ThingDB to a Proto Thing
func DomainThingToProto(t *ThingDB) *thingv1.Thing {
	if t == nil {
		return nil
	}
	return &thingv1.Thing{
		Id:                t.Id,
		TypeId:            t.TypeId,
		Name:              t.Name,
		Description:       derefString(t.Description),
		Comment:           derefString(t.Comment),
		ExternalId:        derefInt32(t.ExternalId),
		ExternalRef:       derefString(t.ExternalRef),
		BuildAt:           timeToTimestamp(t.BuildAt),
		Status:            statusToString(t.Status),
		ContainedBy:       derefString(t.ContainedBy),
		ContainedByOld:    derefInt32(t.ContainedByOld),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timeToTimestamp(t.InactivatedTime),
		InactivatedBy:     derefInt32(t.InactivatedBy),
		InactivatedReason: derefString(t.InactivatedReason),
		Validated:         derefBool(t.Validated),
		ValidatedTime:     timeToTimestamp(t.ValidatedTime),
		ValidatedBy:       derefInt32(t.ValidatedBy),
		ManagedBy:         derefInt32(t.ManagedBy),
		CreatedAt:         timeToTimestamp(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timeToTimestamp(t.LastModifiedAt),
		LastModifiedBy:    derefInt32(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timeToTimestamp(t.DeletedAt),
		DeletedBy:         derefInt32(t.DeletedBy),
		MoreData:          mapToStruct(t.MoreData),
		PosX:              t.PosX,
		PosY:              t.PosY,
	}
}

// ProtoThingToDomain converts a Proto Thing to a domain ThingDB.
func ProtoThingToDomain(t *thingv1.Thing) *ThingDB {
	if t == nil {
		return nil
	}

	return &ThingDB{
		Id:                t.Id,
		TypeId:            t.TypeId,
		Name:              t.Name,
		Description:       stringPtr(t.Description),
		Comment:           stringPtr(t.Comment),
		ExternalId:        int32Ptr(t.ExternalId),
		ExternalRef:       stringPtr(t.ExternalRef),
		BuildAt:           timestampToTime(t.BuildAt),
		Status:            stringToStatus(t.Status),
		ContainedBy:       stringPtr(t.ContainedBy),
		ContainedByOld:    int32Ptr(t.ContainedByOld),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timestampToTime(t.InactivatedTime),
		InactivatedBy:     int32Ptr(t.InactivatedBy),
		InactivatedReason: stringPtr(t.InactivatedReason),
		Validated:         boolPtr(t.Validated),
		ValidatedTime:     timestampToTime(t.ValidatedTime),
		ValidatedBy:       int32Ptr(t.ValidatedBy),
		ManagedBy:         int32Ptr(t.ManagedBy),
		CreatedAt:         timestampToTime(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timestampToTime(t.LastModifiedAt),
		LastModifiedBy:    int32Ptr(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timestampToTime(t.DeletedAt),
		DeletedBy:         int32Ptr(t.DeletedBy),
		MoreData:          structToMap(t.MoreData),
		PosX:              t.PosX,
		PosY:              t.PosY,
	}
}

// DomainThingListToProto converts a domain ThingListDB to a Proto ThingList
func DomainThingListToProto(t *ThingListDB) *thingv1.ThingList {
	if t == nil {
		return nil
	}
	return &thingv1.ThingList{
		Id:          t.Id,
		TypeId:      t.TypeId,
		Name:        t.Name,
		Description: derefString(t.Description),
		ExternalId:  derefInt32(t.ExternalId),
		Inactivated: t.Inactivated,
		Validated:   derefBool(t.Validated),
		Status:      statusToString(t.Status),
		CreatedBy:   t.CreatedBy,
		CreatedAt:   timeToTimestamp(t.CreatedAt),
		PosX:        t.PosX,
		PosY:        t.PosY,
	}
}

// DomainThingListSliceToProto converts a slice of domain ThingListDB to Proto ThingList
func DomainThingListSliceToProto(items []*ThingListDB) []*thingv1.ThingList {
	if items == nil {
		return nil
	}
	result := make([]*thingv1.ThingList, len(items))
	for i, item := range items {
		result[i] = DomainThingListToProto(item)
	}
	return result
}

// =============================================================================
// TypeThing Mappers
// =============================================================================

// DomainTypeThingToProto converts a domain TypeThingDB to a Proto TypeThing
func DomainTypeThingToProto(t *TypeThingDB) *thingv1.TypeThing {
	if t == nil {
		return nil
	}
	return &thingv1.TypeThing{
		Id:                t.Id,
		Name:              t.Name,
		Description:       derefString(t.Description),
		Comment:           derefString(t.Comment),
		ExternalId:        derefInt32(t.ExternalId),
		TableName:         derefString(t.TableName),
		GeometryType:      derefString(t.GeometryType),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timeToTimestamp(t.InactivatedTime),
		InactivatedBy:     derefInt32(t.InactivatedBy),
		InactivatedReason: derefString(t.InactivatedReason),
		ManagedBy:         derefInt32(t.ManagedBy),
		IconPath:          t.IconPath,
		CreatedAt:         timeToTimestamp(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timeToTimestamp(t.LastModifiedAt),
		LastModifiedBy:    derefInt32(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timeToTimestamp(t.DeletedAt),
		DeletedBy:         derefInt32(t.DeletedBy),
		MoreDataSchema:    mapToStruct(t.MoreDataSchema),
	}
}

// ProtoTypeThingToDomain converts a Proto TypeThing to a domain TypeThingDB
func ProtoTypeThingToDomain(t *thingv1.TypeThing) *TypeThingDB {
	if t == nil {
		return nil
	}
	return &TypeThingDB{
		Id:                t.Id,
		Name:              t.Name,
		Description:       stringPtr(t.Description),
		Comment:           stringPtr(t.Comment),
		ExternalId:        int32Ptr(t.ExternalId),
		TableName:         stringPtr(t.TableName),
		GeometryType:      stringPtr(t.GeometryType),
		Inactivated:       t.Inactivated,
		InactivatedTime:   timestampToTime(t.InactivatedTime),
		InactivatedBy:     int32Ptr(t.InactivatedBy),
		InactivatedReason: stringPtr(t.InactivatedReason),
		ManagedBy:         int32Ptr(t.ManagedBy),
		IconPath:          t.IconPath,
		CreatedAt:         timestampToTime(t.CreatedAt),
		CreatedBy:         t.CreatedBy,
		LastModifiedAt:    timestampToTime(t.LastModifiedAt),
		LastModifiedBy:    int32Ptr(t.LastModifiedBy),
		Deleted:           t.Deleted,
		DeletedAt:         timestampToTime(t.DeletedAt),
		DeletedBy:         int32Ptr(t.DeletedBy),
		MoreDataSchema:    structToMap(t.MoreDataSchema),
	}
}

// DomainTypeThingListToProto converts a domain TypeThingListDB to a Proto TypeThingList
func DomainTypeThingListToProto(t *TypeThingListDB) *thingv1.TypeThingList {
	if t == nil {
		return nil
	}
	return &thingv1.TypeThingList{
		Id:           t.Id,
		Name:         t.Name,
		ExternalId:   derefInt32(t.ExternalId),
		IconPath:     t.IconPath,
		CreatedAt:    timeToTimestamp(&t.CreatedAt),
		TableName:    derefString(t.TableName),
		GeometryType: derefString(t.GeometryType),
		Inactivated:  t.Inactivated,
	}
}

// DomainTypeThingListSliceToProto converts a slice of domain TypeThingListDB to Proto
func DomainTypeThingListSliceToProto(items []*TypeThingListDB) []*thingv1.TypeThingList {
	if items == nil {
		return nil
	}
	result := make([]*thingv1.TypeThingList, len(items))
	for i, item := range items {
		result[i] = DomainTypeThingListToProto(item)
	}
	return result
}
