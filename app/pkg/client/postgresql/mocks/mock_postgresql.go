// Code generated by MockGen. DO NOT EDIT.
// Source: utils.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	"skeleton-golange-application/app/model"

	gomock "github.com/golang/mock/gomock"
)

type MockPgClient struct {
	CollectionQuery MockPostgresCollectionQuery
}

// MockPostgresCollectionQuery is a mock of PostgresCollectionQuery interface.
type MockPostgresCollectionQuery struct {
	ctrl     *gomock.Controller
	recorder *MockPostgresCollectionQueryMockRecorder
}

// MockPostgresCollectionQueryMockRecorder is the mock recorder for MockPostgresCollectionQuery.
type MockPostgresCollectionQueryMockRecorder struct {
	mock *MockPostgresCollectionQuery
}

// NewMockPostgresCollectionQuery creates a new mock instance.
func NewMockPostgresCollectionQuery(ctrl *gomock.Controller) *MockPostgresCollectionQuery {
	mock := &MockPostgresCollectionQuery{ctrl: ctrl}
	mock.recorder = &MockPostgresCollectionQueryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPostgresCollectionQuery) EXPECT() *MockPostgresCollectionQueryMockRecorder {
	return m.recorder
}

// CreateIssue mocks base method.
func (m *MockPostgresCollectionQuery) CreateIssue(task *model.Track) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateIssue", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateIssue indicates an expected call of CreateIssue.
func (mr *MockPostgresCollectionQueryMockRecorder) CreateIssue(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssue", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).CreateIssue), task)
}

// CreateMany mocks base method.
func (m *MockPostgresCollectionQuery) CreateMany(list []model.Track) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMany", list)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateMany indicates an expected call of CreateMany.
func (mr *MockPostgresCollectionQueryMockRecorder) CreateMany(list interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMany", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).CreateMany), list)
}

// CreateUser mocks base method.
func (m *MockPostgresCollectionQuery) CreateUser(user model.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", user)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockPostgresCollectionQueryMockRecorder) CreateUser(user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).CreateUser), user)
}

// DeleteAll mocks base method.
func (m *MockPostgresCollectionQuery) DeleteAll() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAll")
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockPostgresCollectionQueryMockRecorder) DeleteAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).DeleteAll))
}

// DeleteOne mocks base method.
func (m *MockPostgresCollectionQuery) DeleteOne(code string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteOne", code)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOne indicates an expected call of DeleteOne.
func (mr *MockPostgresCollectionQueryMockRecorder) DeleteOne(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOne", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).DeleteOne), code)
}

// DeleteUser mocks base method.
func (m *MockPostgresCollectionQuery) DeleteUser(email string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", email)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockPostgresCollectionQueryMockRecorder) DeleteUser(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).DeleteUser), email)
}

// FindUserToEmail mocks base method.
func (m *MockPostgresCollectionQuery) FindUserToEmail(email string) (model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindUserToEmail", email)
	ret0, _ := ret[0].(model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindUserToEmail indicates an expected call of FindUserToEmail.
func (mr *MockPostgresCollectionQueryMockRecorder) FindUserToEmail(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserToEmail", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).FindUserToEmail), email)
}

// GetAllIssues mocks base method.
func (m *MockPostgresCollectionQuery) GetAllIssues() ([]model.Track, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllIssues")
	ret0, _ := ret[0].([]model.Track)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllIssues indicates an expected call of GetAllIssues.
func (mr *MockPostgresCollectionQueryMockRecorder) GetAllIssues() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllIssues", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).GetAllIssues))
}

// GetIssuesByCode mocks base method.
func (m *MockPostgresCollectionQuery) GetIssuesByCode(code string) (model.Track, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIssuesByCode", code)
	ret0, _ := ret[0].(model.Track)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIssuesByCode indicates an expected call of GetIssuesByCode.
func (mr *MockPostgresCollectionQueryMockRecorder) GetIssuesByCode(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssuesByCode", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).GetIssuesByCode), code)
}

// MarkCompleted mocks base method.
func (m *MockPostgresCollectionQuery) MarkCompleted(code string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkCompleted", code)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkCompleted indicates an expected call of MarkCompleted.
func (mr *MockPostgresCollectionQueryMockRecorder) MarkCompleted(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkCompleted", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).MarkCompleted), code)
}

// UpdateIssue mocks base method.
func (m *MockPostgresCollectionQuery) UpdateIssue(track *model.Track) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateIssue", track)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateIssue indicates an expected call of UpdateIssue.
func (mr *MockPostgresCollectionQueryMockRecorder) UpdateIssue(track interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateIssue", reflect.TypeOf((*MockPostgresCollectionQuery)(nil).UpdateIssue), track)
}

// MockPostgresOperations is a mock of PostgresOperations interface.
type MockPostgresOperations struct {
	MockPostgresCollectionQuery
	ctrl     *gomock.Controller
	recorder *MockPostgresOperationsMockRecorder
}

// MockPostgresOperationsMockRecorder is the mock recorder for MockPostgresOperations.
type MockPostgresOperationsMockRecorder struct {
	mock *MockPostgresOperations
}

// NewMockPostgresOperations creates a new mock instance.
func NewMockPostgresOperations(ctrl *gomock.Controller) *MockPostgresOperations {
	mock := &MockPostgresOperations{ctrl: ctrl}
	mock.recorder = &MockPostgresOperationsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPostgresOperations) EXPECT() *MockPostgresOperationsMockRecorder {
	return m.recorder
}

// CreateIssue indicates an expected call of CreateIssue.
func (mr *MockPostgresOperationsMockRecorder) CreateIssue(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIssue", reflect.TypeOf((*MockPostgresOperations)(nil).CreateIssue), task)
}

// CreateMany indicates an expected call of CreateMany.
func (mr *MockPostgresOperationsMockRecorder) CreateMany(list interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMany", reflect.TypeOf((*MockPostgresOperations)(nil).CreateMany), list)
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockPostgresOperationsMockRecorder) DeleteAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockPostgresOperations)(nil).DeleteAll))
}

// DeleteOne indicates an expected call of DeleteOne.
func (mr *MockPostgresOperationsMockRecorder) DeleteOne(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOne", reflect.TypeOf((*MockPostgresOperations)(nil).DeleteOne), code)
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockPostgresOperationsMockRecorder) DeleteUser(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockPostgresOperations)(nil).DeleteUser), email)
}

// FindUserToEmail indicates an expected call of FindUserToEmail.
func (mr *MockPostgresOperationsMockRecorder) FindUserToEmail(email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindUserToEmail", reflect.TypeOf((*MockPostgresOperations)(nil).FindUserToEmail), email)
}

// GetAllIssues indicates an expected call of GetAllIssues.
func (mr *MockPostgresOperationsMockRecorder) GetAllIssues() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllIssues", reflect.TypeOf((*MockPostgresOperations)(nil).GetAllIssues))
}

// GetIssuesByCode indicates an expected call of GetIssuesByCode.
func (mr *MockPostgresOperationsMockRecorder) GetIssuesByCode(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIssuesByCode", reflect.TypeOf((*MockPostgresOperations)(nil).GetIssuesByCode), code)
}

// MarkCompleted indicates an expected call of MarkCompleted.
func (mr *MockPostgresOperationsMockRecorder) MarkCompleted(code interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkCompleted", reflect.TypeOf((*MockPostgresOperations)(nil).MarkCompleted), code)
}

// UpdateIssue indicates an expected call of UpdateIssue.
func (mr *MockPostgresOperationsMockRecorder) UpdateIssue(track interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateIssue", reflect.TypeOf((*MockPostgresOperations)(nil).UpdateIssue), track)
}
