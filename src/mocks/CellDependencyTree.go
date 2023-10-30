// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	bbolt "go.etcd.io/bbolt"

	mock "github.com/stretchr/testify/mock"
)

// CellDependencyTree is an autogenerated mock type for the CellDependencyTree type
type CellDependencyTree struct {
	mock.Mock
}

// GetDependants provides a mock function with given fields: tx, sheetId, dependingOnCellId
func (_m *CellDependencyTree) GetDependants(tx *bbolt.Tx, sheetId []byte, dependingOnCellId string) []string {
	ret := _m.Called(tx, sheetId, dependingOnCellId)

	var r0 []string
	if rf, ok := ret.Get(0).(func(*bbolt.Tx, []byte, string) []string); ok {
		r0 = rf(tx, sheetId, dependingOnCellId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}

// SetDependsOn provides a mock function with given fields: tx, sheetId, dependantCellId, dependingOnCellIds
func (_m *CellDependencyTree) SetDependsOn(tx *bbolt.Tx, sheetId []byte, dependantCellId string, dependingOnCellIds []string) error {
	ret := _m.Called(tx, sheetId, dependantCellId, dependingOnCellIds)

	var r0 error
	if rf, ok := ret.Get(0).(func(*bbolt.Tx, []byte, string, []string) error); ok {
		r0 = rf(tx, sheetId, dependantCellId, dependingOnCellIds)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewCellDependencyTree interface {
	mock.TestingT
	Cleanup(func())
}

// NewCellDependencyTree creates a new instance of CellDependencyTree. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewCellDependencyTree(t mockConstructorTestingTNewCellDependencyTree) *CellDependencyTree {
	mock := &CellDependencyTree{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
