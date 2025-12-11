package thing

import (
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GeoJson(offset, limit int, params GeoJsonParams) (string, error) {
	args := m.Called(offset, limit, params)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) List(offset, limit int, params ListParams) ([]*ThingList, error) {
	args := m.Called(offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ThingList), args.Error(1)
}

func (m *MockStorage) ListByExternalId(offset, limit int, externalId int) ([]*ThingList, error) {
	args := m.Called(offset, limit, externalId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ThingList), args.Error(1)
}

func (m *MockStorage) Search(offset, limit int, params SearchParams) ([]*ThingList, error) {
	args := m.Called(offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ThingList), args.Error(1)
}

func (m *MockStorage) Get(id uuid.UUID) (*Thing, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Thing), args.Error(1)
}

func (m *MockStorage) Exist(id uuid.UUID) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func (m *MockStorage) Count(params CountParams) (int32, error) {
	args := m.Called(params)
	return int32(args.Int(0)), args.Error(1)
}

func (m *MockStorage) Create(thing Thing) (*Thing, error) {
	args := m.Called(thing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Thing), args.Error(1)
}

func (m *MockStorage) Update(id uuid.UUID, thing Thing) (*Thing, error) {
	args := m.Called(id, thing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Thing), args.Error(1)
}

func (m *MockStorage) Delete(id uuid.UUID, userId int32) error {
	args := m.Called(id, userId)
	return args.Error(0)
}

func (m *MockStorage) IsThingActive(id uuid.UUID) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func (m *MockStorage) IsUserOwner(id uuid.UUID, userId int32) bool {
	args := m.Called(id, userId)
	return args.Bool(0)
}

func (m *MockStorage) CreateTypeThing(typeThing TypeThing) (*TypeThing, error) {
	args := m.Called(typeThing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeThing), args.Error(1)
}

func (m *MockStorage) UpdateTypeThing(id int32, typeThing TypeThing) (*TypeThing, error) {
	args := m.Called(id, typeThing)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeThing), args.Error(1)
}

func (m *MockStorage) DeleteTypeThing(id int32, userId int32) error {
	args := m.Called(id, userId)
	return args.Error(0)
}

func (m *MockStorage) ListTypeThing(offset, limit int, params TypeThingListParams) ([]*TypeThingList, error) {
	args := m.Called(offset, limit, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*TypeThingList), args.Error(1)
}

func (m *MockStorage) GetTypeThing(id int32) (*TypeThing, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TypeThing), args.Error(1)
}

func (m *MockStorage) CountTypeThing(params TypeThingCountParams) (int32, error) {
	args := m.Called(params)
	return int32(args.Int(0)), args.Error(1)
}

// MockDB is a minimal mock for database connection
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetQueryInt(query string, args ...interface{}) (int, error) {
	callArgs := m.Called(query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) GetVersion() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockDB) Close() {
	m.Called()
}

func (m *MockDB) GetQueryBool(query string, args ...interface{}) (bool, error) {
	callArgs := m.Called(query, args)
	return callArgs.Bool(0), callArgs.Error(1)
}

func (m *MockDB) ExecActionQuery(query string, args ...interface{}) (int, error) {
	callArgs := m.Called(query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

func (m *MockDB) DoesTableExist(schema, table string) bool {
	args := m.Called(schema, table)
	return args.Bool(0)
}

func (m *MockDB) GetPGConn() (*pgxpool.Pool, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pgxpool.Pool), args.Error(1)
}

func (m *MockDB) GetQueryString(query string, args ...interface{}) (string, error) {
	callArgs := m.Called(query, args)
	return callArgs.String(0), callArgs.Error(1)
}

func (m *MockDB) Insert(query string, args ...interface{}) (int, error) {
	callArgs := m.Called(query, args)
	return callArgs.Int(0), callArgs.Error(1)
}

// Helper function to create a test business service
func createTestBusinessService(mockStore *MockStorage, mockDB *MockDB) *BusinessService {
	logger, _ := golog.NewLogger("simple", os.Stdout, golog.InfoLevel, "test")
	return NewBusinessService(mockStore, mockDB, logger, 50)
}

// Test Create operation
func TestBusinessService_Create(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		newThing := Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		expectedThing := newThing
		expectedThing.CreatedBy = 123

		mockStore.On("Exist", thingID).Return(false)
		mockStore.On("Create", mock.AnythingOfType("Thing")).Return(&expectedThing, nil)

		result, err := service.Create(123, newThing)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation error - empty name", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newThing := Thing{
			Id:   uuid.New(),
			Name: "  ", // Empty/whitespace name
		}

		result, err := service.Create(123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("validation error - name too short", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newThing := Thing{
			Id:   uuid.New(),
			Name: "ab", // Less than MinNameLength (5)
		}

		result, err := service.Create(123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("already exists error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		newThing := Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		mockStore.On("Exist", thingID).Return(true)

		result, err := service.Create(123, newThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		mockStore.AssertExpectations(t)
	})
}

// Test Get operation
func TestBusinessService_Get(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		expectedThing := &Thing{
			Id:   thingID,
			Name: "Test Thing",
		}

		mockStore.On("Exist", thingID).Return(true)
		mockStore.On("Get", thingID).Return(expectedThing, nil)

		result, err := service.Get(thingID)

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
		mockStore.On("Exist", thingID).Return(false)

		result, err := service.Get(thingID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrNotFound)
		mockStore.AssertExpectations(t)
	})
}

// Test Update operation
func TestBusinessService_Update(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)
		updateThing := Thing{
			Id:   thingID,
			Name: "Updated Thing",
		}

		expectedThing := updateThing
		expectedThing.LastModifiedBy = &userID

		mockStore.On("Exist", thingID).Return(true)
		mockStore.On("IsUserOwner", thingID, userID).Return(true)
		mockStore.On("Update", thingID, mock.AnythingOfType("Thing")).Return(&expectedThing, nil)

		result, err := service.Update(userID, thingID, updateThing)

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
		updateThing := Thing{
			Id:   thingID,
			Name: "Updated Thing",
		}

		mockStore.On("Exist", thingID).Return(true)
		mockStore.On("IsUserOwner", thingID, userID).Return(false)

		result, err := service.Update(userID, thingID, updateThing)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
}

// Test Delete operation
func TestBusinessService_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", thingID).Return(true)
		mockStore.On("IsUserOwner", thingID, userID).Return(true)
		mockStore.On("Delete", thingID, userID).Return(nil)

		err := service.Delete(userID, thingID)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("not owner error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		thingID := uuid.New()
		userID := int32(123)

		mockStore.On("Exist", thingID).Return(true)
		mockStore.On("IsUserOwner", thingID, userID).Return(false)

		err := service.Delete(userID, thingID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnauthorized)
		mockStore.AssertExpectations(t)
	})
}

// Test List operation
func TestBusinessService_List(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		expectedList := []*ThingList{
			{Id: uuid.New(), Name: "Thing 1"},
			{Id: uuid.New(), Name: "Thing 2"},
		}
		params := ListParams{}

		mockStore.On("List", 0, 10, params).Return(expectedList, nil)

		result, err := service.List(0, 10, params)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockStore.AssertExpectations(t)
	})

	t.Run("empty list with pgx.ErrNoRows", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := ListParams{}
		mockStore.On("List", 0, 10, params).Return(nil, pgx.ErrNoRows)

		result, err := service.List(0, 10, params)

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockStore.AssertExpectations(t)
	})

	t.Run("database error", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := ListParams{}
		dbError := errors.New("database connection failed")
		mockStore.On("List", 0, 10, params).Return(nil, dbError)

		result, err := service.List(0, 10, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
	})
}

// Test Count operation
func TestBusinessService_Count(t *testing.T) {
	t.Run("successful count", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		params := CountParams{}
		mockStore.On("Count", params).Return(42, nil)

		result, err := service.Count(params)

		assert.NoError(t, err)
		assert.Equal(t, int32(42), result)
		mockStore.AssertExpectations(t)
	})
}

// Test CreateTypeThing operation
func TestBusinessService_CreateTypeThing(t *testing.T) {
	t.Run("successful creation by admin", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeThing := TypeThing{
			Name: "Test Type",
		}

		expectedTypeThing := newTypeThing
		expectedTypeThing.Id = 1
		expectedTypeThing.CreatedBy = 123

		mockStore.On("CreateTypeThing", mock.AnythingOfType("TypeThing")).Return(&expectedTypeThing, nil)

		result, err := service.CreateTypeThing(123, true, newTypeThing)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int32(123), result.CreatedBy)
		mockStore.AssertExpectations(t)
	})

	t.Run("non-admin rejection", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockDB := new(MockDB)
		service := createTestBusinessService(mockStore, mockDB)

		newTypeThing := TypeThing{
			Name: "Test Type",
		}

		result, err := service.CreateTypeThing(123, false, newTypeThing)

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
