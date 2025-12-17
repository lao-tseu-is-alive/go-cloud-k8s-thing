package thing

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// =============================================================================
// Test Helpers
// =============================================================================

// Helper to create a test Connect server
func createTestThingConnectServer(mockStore *MockStorage, mockDB *MockDB) *ThingConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewThingConnectServer(businessService, logger)
}

// Helper to create a test TypeThing Connect server
func createTestTypeThingConnectServer(mockStore *MockStorage, mockDB *MockDB) *TypeThingConnectServer {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewTypeThingConnectServer(businessService, logger)
}

// Helper to create a context with user info (simulating what AuthInterceptor does)
func contextWithUser(userId int32, isAdmin bool) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, userIDKey, userId)
	ctx = context.WithValue(ctx, isAdminKey, isAdmin)
	return ctx
}

// Helper to create a Connect request (no auth header needed since we inject via context)
func createConnectRequest[T any](msg *T) *connect.Request[T] {
	return connect.NewRequest(msg)
}

// =============================================================================
// ThingConnectServer Tests
// =============================================================================

func TestThingConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		// Setup mock storage
		now := time.Now()
		expectedList := []*ThingList{
			{Id: uuid.New(), Name: "Thing 1", CreatedAt: &now},
			{Id: uuid.New(), Name: "Thing 2", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anything, 0, 50, ListParams{}).Return(expectedList, nil)

		// Create request and context with user
		req := createConnectRequest(&thingv1.ListRequest{Limit: 0, Offset: 0})
		ctx := contextWithUser(123, false)

		// Call handler
		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Things, 2)
		assert.Equal(t, "Thing 1", resp.Msg.Things[0].Name)
		assert.Equal(t, "Thing 2", resp.Msg.Things[1].Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("list with pagination", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*ThingList{
			{Id: uuid.New(), Name: "Thing 3", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anything, 10, 5, ListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&thingv1.ListRequest{Limit: 5, Offset: 10})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Things, 1)
		mockStore.AssertExpectations(t)
	})
}

func TestThingConnectServer_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		thingID := uuid.New()
		expectedThing := &Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("Get", mock.Anything, thingID).Return(expectedThing, nil)

		req := createConnectRequest(&thingv1.GetRequest{Id: thingID.String()})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, thingID.String(), resp.Msg.Thing.Id)
		assert.Equal(t, "Test Thing", resp.Msg.Thing.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		thingID := uuid.New()

		mockStore.On("Exist", mock.Anything, thingID).Return(false)

		req := createConnectRequest(&thingv1.GetRequest{Id: thingID.String()})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeNotFound, connectErr.Code())
		mockStore.AssertExpectations(t)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		req := createConnectRequest(&thingv1.GetRequest{Id: "not-a-uuid"})
		ctx := contextWithUser(123, false)

		resp, err := server.Get(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestThingConnectServer_Create(t *testing.T) {
	t.Run("successful create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		thingID := uuid.New()
		expectedThing := &Thing{
			Id:        thingID,
			Name:      "New Thing",
			CreatedBy: 123,
		}

		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, mock.Anything).Return(1, nil)
		mockStore.On("Exist", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(false)
		mockStore.On("Create", mock.Anything, mock.AnythingOfType("Thing")).Return(expectedThing, nil)

		req := createConnectRequest(&thingv1.CreateRequest{
			Thing: &thingv1.Thing{
				Name: "New Thing",
			},
		})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Thing", resp.Msg.Thing.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - missing thing", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		req := createConnectRequest(&thingv1.CreateRequest{Thing: nil})
		ctx := contextWithUser(123, false)

		resp, err := server.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	})
}

func TestThingConnectServer_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(true)
		mockStore.On("Delete", mock.Anything, thingID, userID).Return(nil)

		req := createConnectRequest(&thingv1.DeleteRequest{Id: thingID.String()})
		ctx := contextWithUser(userID, false)

		resp, err := server.Delete(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockStore.AssertExpectations(t)
	})

	t.Run("permission denied - not owner", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(false)

		req := createConnectRequest(&thingv1.DeleteRequest{Id: thingID.String()})
		ctx := contextWithUser(userID, false)

		resp, err := server.Delete(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		mockStore.AssertExpectations(t)
	})
}

func TestThingConnectServer_Count(t *testing.T) {
	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestThingConnectServer(mockStore, mockDB)

		mockStore.On("Count", mock.Anything, CountParams{}).Return(42, nil)

		req := createConnectRequest(&thingv1.CountRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.Count(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(42), resp.Msg.Count)
		mockStore.AssertExpectations(t)
	})
}

// =============================================================================
// TypeThingConnectServer Tests
// =============================================================================

func TestTypeThingConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeThingConnectServer(mockStore, mockDB)

		now := time.Now()
		expectedList := []*TypeThingList{
			{Id: 1, Name: "Type 1", CreatedAt: now},
			{Id: 2, Name: "Type 2", CreatedAt: now},
		}

		mockStore.On("ListTypeThing", mock.Anything, 0, 250, TypeThingListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&thingv1.TypeThingListRequest{})
		ctx := contextWithUser(123, false)

		resp, err := server.List(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.TypeThings, 2)
		mockStore.AssertExpectations(t)
	})
}

func TestTypeThingConnectServer_Create(t *testing.T) {
	t.Run("admin can create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeThingConnectServer(mockStore, mockDB)

		expectedTypeThing := &TypeThing{
			Id:        1,
			Name:      "New Type",
			CreatedBy: 123,
		}

		mockStore.On("CreateTypeThing", mock.Anything, mock.AnythingOfType("TypeThing")).Return(expectedTypeThing, nil)

		req := createConnectRequest(&thingv1.TypeThingCreateRequest{
			TypeThing: &thingv1.TypeThing{
				Name: "New Type",
			},
		})
		ctx := contextWithUser(123, true) // isAdmin = true

		resp, err := server.Create(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Type", resp.Msg.TypeThing.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		server := createTestTypeThingConnectServer(mockStore, mockDB)

		req := createConnectRequest(&thingv1.TypeThingCreateRequest{
			TypeThing: &thingv1.TypeThing{
				Name: "New Type",
			},
		})
		ctx := contextWithUser(123, false) // isAdmin = false

		resp, err := server.Create(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
	})
}

// =============================================================================
// AuthInterceptor Tests
// =============================================================================

func TestGetUserFromContext(t *testing.T) {
	t.Run("user present in context", func(t *testing.T) {
		ctx := contextWithUser(456, true)

		userId, isAdmin := GetUserFromContext(ctx)

		assert.Equal(t, int32(456), userId)
		assert.True(t, isAdmin)
	})

	t.Run("user not present in context", func(t *testing.T) {
		ctx := context.Background()

		userId, isAdmin := GetUserFromContext(ctx)

		assert.Equal(t, int32(0), userId)
		assert.False(t, isAdmin)
	})
}
