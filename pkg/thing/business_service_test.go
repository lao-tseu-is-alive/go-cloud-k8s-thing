package thing

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	thingv1 "github.com/lao-tseu-is-alive/go-cloud-k8s-thing/gen/thing/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GeoJson(ctx context.Context, req *thingv1.GeoJsonRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) List(ctx context.Context, req *thingv1.ListRequest) ([]*thingv1.ThingList, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*thingv1.ThingList), args.Error(1)
}

func (m *MockStorage) ListByExternalId(ctx context.Context, req *thingv1.ListByExternalIdRequest) ([]*thingv1.ThingList, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*thingv1.ThingList), args.Error(1)
}

func (m *MockStorage) Search(ctx context.Context, req *thingv1.SearchRequest) ([]*thingv1.ThingList, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*thingv1.ThingList), args.Error(1)
}

func (m *MockStorage) Get(ctx context.Context, id uuid.UUID) (*thingv1.Thing, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.Thing), args.Error(1)
}

func (m *MockStorage) Exist(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) Count(ctx context.Context, req *thingv1.CountRequest) (int32, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int32), args.Error(1)
}

func (m *MockStorage) Create(ctx context.Context, thing *thingv1.Thing) (*thingv1.Thing, error) {
	args := m.Called(ctx, thing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.Thing), args.Error(1)
}

func (m *MockStorage) Update(ctx context.Context, id uuid.UUID, thing *thingv1.Thing) (*thingv1.Thing, error) {
	args := m.Called(ctx, id, thing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.Thing), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, id uuid.UUID, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) IsThingActive(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) IsUserOwner(ctx context.Context, id uuid.UUID, userId int32) (bool, error) {
	args := m.Called(ctx, id, userId)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) CreateTypeThing(ctx context.Context, typeThing *thingv1.TypeThing) (*thingv1.TypeThing, error) {
	args := m.Called(ctx, typeThing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.TypeThing), args.Error(1)
}

func (m *MockStorage) UpdateTypeThing(ctx context.Context, id int32, typeThing *thingv1.TypeThing) (*thingv1.TypeThing, error) {
	args := m.Called(ctx, id, typeThing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.TypeThing), args.Error(1)
}

func (m *MockStorage) DeleteTypeThing(ctx context.Context, id int32, userId int32) error {
	args := m.Called(ctx, id, userId)
	return args.Error(0)
}

func (m *MockStorage) ListTypeThing(ctx context.Context, req *thingv1.TypeThingServiceListRequest) ([]*thingv1.TypeThingList, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*thingv1.TypeThingList), args.Error(1)
}

func (m *MockStorage) GetTypeThing(ctx context.Context, id int32) (*thingv1.TypeThing, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*thingv1.TypeThing), args.Error(1)
}

func (m *MockStorage) CountTypeThing(ctx context.Context, req *thingv1.TypeThingServiceCountRequest) (int32, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(int32), args.Error(1)
}

// MockDB is a minimal mock for database connection
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetQueryInt(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) GetVersion(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockDB) Close() {
	m.Called()
}

func (m *MockDB) HealthCheck(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockDB) GetQueryBool(ctx context.Context, query string, args ...interface{}) (bool, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Bool(0), callArgs.Error(1)
}

func (m *MockDB) ExecActionQuery(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) DoesTableExist(ctx context.Context, schema, table string) bool {
	args := m.Called(ctx, schema, table)
	return args.Bool(0)
}

func (m *MockDB) GetPGConn() (*pgxpool.Pool, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgxpool.Pool), args.Error(1)
}

func (m *MockDB) GetQueryString(ctx context.Context, query string, args ...interface{}) (string, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.String(0), callArgs.Error(1)
}

func (m *MockDB) Insert(ctx context.Context, query string, args ...interface{}) (int, error) {
	callArgs := m.Called(ctx, query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

// Helper function to create a test business service
func createTestBusinessService(mockStore *MockStorage, mockDB *MockDB) *BusinessService {
	logger := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	return NewBusinessService(mockStore, mockDB, logger, 50)
}

// Test Create operation
func TestBusinessService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New().String()
		newThing := &thingv1.Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		expectedThing := &thingv1.Thing{
			Id:        thingID,
			Name:      "Test Thing",
			CreatedBy: 123,
		}

		// Mock TypeThing existence check
		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, []interface{}{newThing.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anything, mock.Anything).Return(false, nil)
		mockStore.On("Create", mock.Anything, mock.AnythingOfType("*thingv1.Thing")).Return(expectedThing, nil)

		result, err := service.Create(ctx, 123, newThing)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newThing := &thingv1.Thing{
			Id:   uuid.New().String(),
			Name: "  ", // Empty/whitespace name
		}

		result, err := service.Create(ctx, 123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - name too short", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newThing := &thingv1.Thing{
			Id:   uuid.New().String(),
			Name: "ab", // Less than MinNameLength (5)
		}

		result, err := service.Create(ctx, 123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newThing := &thingv1.Thing{
			Id:     uuid.New().String(),
			Name:   "Test Thing",
			TypeId: 999,
		}

		// Mock TypeThing existence check failure
		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, []interface{}{newThing.TypeId}).Return(0, nil)

		result, err := service.Create(ctx, 123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeThingNotFound)
	})

	t.Run("already exists error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New().String()
		newThing := &thingv1.Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		// Mock TypeThing existence check
		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, []interface{}{newThing.TypeId}).Return(1, nil)
		mockStore.On("Exist", mock.Anything, mock.Anything).Return(true, nil)

		result, err := service.Create(ctx, 123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockStore.AssertExpectations(t)
	})
}

// Test Get operation
func TestBusinessService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		expectedThing := &thingv1.Thing{
			Id:   thingID.String(),
			Name: "Test Thing",
		}

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("Get", mock.Anything, thingID).Return(expectedThing, nil)

		result, err := service.Get(ctx, thingID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Thing", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("thing not found", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		mockStore.On("Exist", mock.Anything, thingID).Return(false, nil)

		result, err := service.Get(ctx, thingID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrNotFound)
		mockStore.AssertExpectations(t)
	})
}

// Test Update operation
func TestBusinessService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)
		updateThing := &thingv1.Thing{
			Id:   thingID.String(),
			Name: "Updated Thing",
		}

		expectedThing := &thingv1.Thing{
			Id:             thingID.String(),
			Name:           "Updated Thing",
			LastModifiedBy: userID,
		}

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(true, nil)
		// Mock TypeThing existence check
		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, []interface{}{updateThing.TypeId}).Return(1, nil)
		mockStore.On("Update", mock.Anything, thingID, mock.AnythingOfType("*thingv1.Thing")).Return(expectedThing, nil)

		result, err := service.Update(ctx, userID, thingID, updateThing)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Thing", result.Name)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)
		updateThing := &thingv1.Thing{
			Id:   thingID.String(),
			Name: "Updated Thing",
		}

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(false, nil)

		result, err := service.Update(ctx, userID, thingID, updateThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
	t.Run("validation error - invalid type id", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)
		updateThing := &thingv1.Thing{
			Id:     thingID.String(),
			Name:   "Updated Thing",
			TypeId: 999,
		}

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(true, nil)
		// Mock TypeThing existence check failure
		mockDB.On("GetQueryInt", mock.Anything, existTypeThing, []interface{}{updateThing.TypeId}).Return(0, nil)

		result, err := service.Update(ctx, userID, thingID, updateThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrTypeThingNotFound)
		mockStore.AssertExpectations(t)
	})
}

// Test Delete operation
func TestBusinessService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(true, nil)
		mockStore.On("Delete", mock.Anything, thingID, userID).Return(nil)

		err := service.Delete(ctx, userID, thingID)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", mock.Anything, thingID).Return(true, nil)
		mockStore.On("IsUserOwner", mock.Anything, thingID, userID).Return(false, nil)

		err := service.Delete(ctx, userID, thingID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
}

// Test List operation
func TestBusinessService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		expectedList := []*thingv1.ThingList{
			{Id: uuid.New().String(), Name: "Thing 1"},
			{Id: uuid.New().String(), Name: "Thing 2"},
		}
		req := &thingv1.ListRequest{Limit: 10, Offset: 0}

		mockStore.On("List", mock.Anything, req).Return(expectedList, nil)

		result, err := service.List(ctx, req)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty list with ErrEmptyResult", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		req := &thingv1.ListRequest{Limit: 10, Offset: 0}
		mockStore.On("List", mock.Anything, req).Return(nil, ErrEmptyResult)

		result, err := service.List(ctx, req)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockStore.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		req := &thingv1.ListRequest{Limit: 10, Offset: 0}
		dbError := errors.New("database connection failed")
		mockStore.On("List", mock.Anything, req).Return(nil, dbError)

		result, err := service.List(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
	})
}

// Test Count operation
func TestBusinessService_Count(t *testing.T) {
	ctx := context.Background()

	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		req := &thingv1.CountRequest{}
		mockStore.On("Count", mock.Anything, req).Return(int32(42), nil)

		result, err := service.Count(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, int32(42), result)
		mockStore.AssertExpectations(t)
	})
}

// Test CreateTypeThing operation
func TestBusinessService_CreateTypeThing(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation by admin", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeThing := &thingv1.TypeThing{
			Name: "Test Type",
		}

		expectedTypeThing := &thingv1.TypeThing{
			Name:      "Test Type",
			Id:        1,
			CreatedBy: 123,
		}

		mockStore.On("CreateTypeThing", mock.Anything, mock.AnythingOfType("*thingv1.TypeThing")).Return(expectedTypeThing, nil)

		result, err := service.CreateTypeThing(ctx, 123, true, newTypeThing)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejection", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeThing := &thingv1.TypeThing{
			Name: "Test Type",
		}

		result, err := service.CreateTypeThing(ctx, 123, false, newTypeThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAdminRequired)
	})
}

// Test validation function
func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{"valid name", "Valid Name", false},
		{"empty string", "", true},
		{"only spaces", "   ", true},
		{"too short", "ab", true},
		{"exactly min length", "12345", false},
		{"longer than min", "Long Enough Name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
