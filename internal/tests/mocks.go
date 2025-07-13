package tests

//
//import (
//	"github.com/stretchr/testify/mock"
//
//	"github.com/alisaviation/internal/gophermart/models"
//)
//
//type MockOrderDB struct {
//	mock.Mock
//}
//
//func (m *MockOrderDB) GetOrdersByUser(userID int) ([]models.Order, error) {
//	args := m.Called(userID)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).([]models.Order), args.Error(1)
//}
//
//func (m *MockOrderDB) GetOrderByNumber(number string) (*models.Order, error) {
//	args := m.Called(number)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*models.Order), args.Error(1)
//}
//
//func (m *MockOrderDB) CreateOrder(order *models.Order) error {
//	args := m.Called(order)
//	return args.Error(0)
//}
//
//type MockUserRepository struct {
//	GetUserByLoginFunc func(login string) (*models.User, error)
//	CreateUserFunc     func(user models.User) (int, error)
//}
//
//func (m *MockUserRepository) GetUserByLogin(login string) (*models.User, error) {
//	return m.GetUserByLoginFunc(login)
//}
//
//func (m *MockUserRepository) CreateUser(user models.User) (int, error) {
//	return m.CreateUserFunc(user)
//}
//
//type MockJWTService struct {
//	GenerateTokenFunc func(userID int, login string) (string, error)
//}
//
//func (m *MockJWTService) GenerateToken(userID int, login string) (string, error) {
//	return m.GenerateTokenFunc(userID, login)
//}
