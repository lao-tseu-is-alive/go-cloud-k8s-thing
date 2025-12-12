package thing

import (
	"context"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/cristalhq/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/goHttpEcho"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJwtChecker implements goHttpEcho.JwtChecker for testing
type MockJwtChecker struct {
	mock.Mock
}

func (m *MockJwtChecker) ParseToken(jwtToken string) (*goHttpEcho.JwtCustomClaims, error) {
	args := m.Called(jwtToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*goHttpEcho.JwtCustomClaims), args.Error(1)
}

func (m *MockJwtChecker) GetTokenFromUserInfo(userInfo *goHttpEcho.UserInfo) (*jwt.Token, error) {
	args := m.Called(userInfo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJwtChecker) JwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return nil
}

func (m *MockJwtChecker) GetLogger() golog.MyLogger {
	return nil
}

func (m *MockJwtChecker) GetJwtDuration() int {
	return 3600
}

func (m *MockJwtChecker) GetIssuerId() string {
	return "test-issuer"
}

func (m *MockJwtChecker) GetJwtCustomClaimsFromContext(c echo.Context) *goHttpEcho.JwtCustomClaims {
	return nil
}

// Helper to create a test Connect server
func createTestThingConnectServer(mockStore *MockStorage, mockDB *MockDB, mockJwt *MockJwtChecker) *ThingConnectServer {
	logger, _ := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewThingConnectServer(businessService, logger, mockJwt)
}

// Helper to create a Connect request with Authorization header
func createConnectRequest[T any](msg *T, token string) *connect.Request[T] {
	req := connect.NewRequest(msg)
	req.Header().Set("Authorization", "Bearer "+token)
	return req
}

// Helper to create mock JWT claims
func createMockClaims(userId int, isAdmin bool) *goHttpEcho.JwtCustomClaims {
	return &goHttpEcho.JwtCustomClaims{
		User: &goHttpEcho.UserInfo{
			UserId:  userId,
			IsAdmin: isAdmin,
			Name:    "Test User",
			Email:   "test@example.com",
		},
	}
}

// =============================================================================
// ThingConnectServer Tests
// =============================================================================

func TestThingConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		// Setup mock JWT validation
		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)

		// Setup mock storage
		now := time.Now()
		expectedList := []*ThingList{
			{Id: uuid.New(), Name: "Thing 1", CreatedAt: &now},
			{Id: uuid.New(), Name: "Thing 2", CreatedAt: &now},
		}
		mockStore.On("List", mock.Anything, 0, 50, ListParams{}).Return(expectedList, nil)

		// Create request
		req := createConnectRequest(&thingv1.ListRequest{Limit: 0, Offset: 0}, "valid-token")

		// Call handler
		resp, err := server.List(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.Things, 2)
		assert.Equal(t, "Thing 1", resp.Msg.Things[0].Name)
		assert.Equal(t, "Thing 2", resp.Msg.Things[1].Name)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})

	t.Run("unauthorized - missing token", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		// Create request without Authorization header
		req := connect.NewRequest(&thingv1.ListRequest{})

		// Call handler
		resp, err := server.List(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
	})

	t.Run("unauthorized - invalid token", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		// Setup mock JWT validation to fail
		mockJwt.On("ParseToken", "invalid-token").Return(nil, assert.AnError)

		// Create request with invalid token
		req := createConnectRequest(&thingv1.ListRequest{}, "invalid-token")

		// Call handler
		resp, err := server.List(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
		mockJwt.AssertExpectations(t)
	})
}

func TestThingConnectServer_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		thingID := uuid.New()
		expectedThing := &Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)
		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("Get", mock.Anything, thingID).Return(expectedThing, nil)

		req := createConnectRequest(&thingv1.GetRequest{Id: thingID.String()}, "valid-token")

		resp, err := server.Get(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, thingID.String(), resp.Msg.Thing.Id)
		assert.Equal(t, "Test Thing", resp.Msg.Thing.Name)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		thingID := uuid.New()

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)
		mockStore.On("Exist", mock.Anything, thingID).Return(false)

		req := createConnectRequest(&thingv1.GetRequest{Id: thingID.String()}, "valid-token")

		resp, err := server.Get(context.Background(), req)

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
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)

		req := createConnectRequest(&thingv1.GetRequest{Id: "not-a-uuid"}, "valid-token")

		resp, err := server.Get(context.Background(), req)

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
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		thingID := uuid.New()
		expectedThing := &Thing{
			Id:        thingID,
			Name:      "New Thing",
			CreatedBy: 123,
		}

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)
		mockStore.On("Exist", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(false)
		mockStore.On("Create", mock.Anything, mock.AnythingOfType("Thing")).Return(expectedThing, nil)

		req := createConnectRequest(&thingv1.CreateRequest{
			Thing: &thingv1.Thing{
				Name: "New Thing",
			},
		}, "valid-token")

		resp, err := server.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Thing", resp.Msg.Thing.Name)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})

	t.Run("validation error - missing thing", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)

		req := createConnectRequest(&thingv1.CreateRequest{Thing: nil}, "valid-token")

		resp, err := server.Create(context.Background(), req)

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
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		thingID := uuid.New()
		userID := int32(123)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(int(userID), false), nil)
		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(true)
		mockStore.On("Delete", mock.Anything, thingID, userID).Return(nil)

		req := createConnectRequest(&thingv1.DeleteRequest{Id: thingID.String()}, "valid-token")

		resp, err := server.Delete(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})

	t.Run("permission denied - not owner", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		thingID := uuid.New()
		userID := int32(123)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(int(userID), false), nil)
		mockStore.On("Exist", mock.Anything, thingID).Return(true)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(false)

		req := createConnectRequest(&thingv1.DeleteRequest{Id: thingID.String()}, "valid-token")

		resp, err := server.Delete(context.Background(), req)

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
		mockJwt := new(MockJwtChecker)
		server := createTestThingConnectServer(mockStore, mockDB, mockJwt)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)
		mockStore.On("Count", mock.Anything, CountParams{}).Return(42, nil)

		req := createConnectRequest(&thingv1.CountRequest{}, "valid-token")

		resp, err := server.Count(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int32(42), resp.Msg.Count)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})
}

// =============================================================================
// TypeThingConnectServer Tests (for completeness)
// =============================================================================

func createTestTypeThingConnectServer(mockStore *MockStorage, mockDB *MockDB, mockJwt *MockJwtChecker) *TypeThingConnectServer {
	logger, _ := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	businessService := NewBusinessService(mockStore, mockDB, logger, 50)
	return NewTypeThingConnectServer(businessService, logger, mockJwt)
}

func TestTypeThingConnectServer_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestTypeThingConnectServer(mockStore, mockDB, mockJwt)

		now := time.Now()
		expectedList := []*TypeThingList{
			{Id: 1, Name: "Type 1", CreatedAt: now},
			{Id: 2, Name: "Type 2", CreatedAt: now},
		}

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)
		mockStore.On("ListTypeThing", mock.Anything, 0, 250, TypeThingListParams{}).Return(expectedList, nil)

		req := createConnectRequest(&thingv1.TypeThingListRequest{}, "valid-token")

		resp, err := server.List(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Msg.TypeThings, 2)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})
}

func TestTypeThingConnectServer_Create(t *testing.T) {
	t.Run("admin can create", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestTypeThingConnectServer(mockStore, mockDB, mockJwt)

		expectedTypeThing := &TypeThing{
			Id:        1,
			Name:      "New Type",
			CreatedBy: 123,
		}

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, true), nil)
		mockStore.On("CreateTypeThing", mock.Anything, mock.AnythingOfType("TypeThing")).Return(expectedTypeThing, nil)

		req := createConnectRequest(&thingv1.TypeThingCreateRequest{
			TypeThing: &thingv1.TypeThing{
				Name: "New Type",
			},
		}, "valid-token")

		resp, err := server.Create(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Type", resp.Msg.TypeThing.Name)
		mockStore.AssertExpectations(t)
		mockJwt.AssertExpectations(t)
	})

	t.Run("non-admin rejected", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		mockJwt := new(MockJwtChecker)
		server := createTestTypeThingConnectServer(mockStore, mockDB, mockJwt)

		mockJwt.On("ParseToken", "valid-token").Return(createMockClaims(123, false), nil)

		req := createConnectRequest(&thingv1.TypeThingCreateRequest{
			TypeThing: &thingv1.TypeThing{
				Name: "New Type",
			},
		}, "valid-token")

		resp, err := server.Create(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		mockJwt.AssertExpectations(t)
	})
}
