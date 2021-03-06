// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/whatthefar/monorepo-toolkit/pkg/core (interfaces: PipelineGateway)

// Package mock_core is a generated GoMock package.
package mock_core

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	core "github.com/whatthefar/monorepo-toolkit/pkg/core"
	reflect "reflect"
)

// MockPipelineGateway is a mock of PipelineGateway interface
type MockPipelineGateway struct {
	ctrl     *gomock.Controller
	recorder *MockPipelineGatewayMockRecorder
}

// MockPipelineGatewayMockRecorder is the mock recorder for MockPipelineGateway
type MockPipelineGatewayMockRecorder struct {
	mock *MockPipelineGateway
}

// NewMockPipelineGateway creates a new mock instance
func NewMockPipelineGateway(ctrl *gomock.Controller) *MockPipelineGateway {
	mock := &MockPipelineGateway{ctrl: ctrl}
	mock.recorder = &MockPipelineGatewayMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPipelineGateway) EXPECT() *MockPipelineGatewayMockRecorder {
	return m.recorder
}

// BuildStatus mocks base method
func (m *MockPipelineGateway) BuildStatus(arg0 context.Context, arg1 string) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildStatus", arg0, arg1)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildStatus indicates an expected call of BuildStatus
func (mr *MockPipelineGatewayMockRecorder) BuildStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildStatus", reflect.TypeOf((*MockPipelineGateway)(nil).BuildStatus), arg0, arg1)
}

// CurrentCommit mocks base method
func (m *MockPipelineGateway) CurrentCommit() core.Hash {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentCommit")
	ret0, _ := ret[0].(core.Hash)
	return ret0
}

// CurrentCommit indicates an expected call of CurrentCommit
func (mr *MockPipelineGatewayMockRecorder) CurrentCommit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentCommit", reflect.TypeOf((*MockPipelineGateway)(nil).CurrentCommit))
}

// KillBuild mocks base method
func (m *MockPipelineGateway) KillBuild(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KillBuild", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// KillBuild indicates an expected call of KillBuild
func (mr *MockPipelineGatewayMockRecorder) KillBuild(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KillBuild", reflect.TypeOf((*MockPipelineGateway)(nil).KillBuild), arg0, arg1)
}

// LastSuccessfulCommit mocks base method
func (m *MockPipelineGateway) LastSuccessfulCommit(arg0 context.Context, arg1 string) (core.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LastSuccessfulCommit", arg0, arg1)
	ret0, _ := ret[0].(core.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LastSuccessfulCommit indicates an expected call of LastSuccessfulCommit
func (mr *MockPipelineGatewayMockRecorder) LastSuccessfulCommit(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LastSuccessfulCommit", reflect.TypeOf((*MockPipelineGateway)(nil).LastSuccessfulCommit), arg0, arg1)
}

// TriggerBuild mocks base method
func (m *MockPipelineGateway) TriggerBuild(arg0 context.Context, arg1 string) (*string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TriggerBuild", arg0, arg1)
	ret0, _ := ret[0].(*string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TriggerBuild indicates an expected call of TriggerBuild
func (mr *MockPipelineGatewayMockRecorder) TriggerBuild(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TriggerBuild", reflect.TypeOf((*MockPipelineGateway)(nil).TriggerBuild), arg0, arg1)
}
