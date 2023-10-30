// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	gin "github.com/gin-gonic/gin"
	mock "github.com/stretchr/testify/mock"
)

// ApiController is an autogenerated mock type for the ApiController type
type ApiController struct {
	mock.Mock
}

// GetCellAction provides a mock function with given fields: c
func (_m *ApiController) GetCellAction(c *gin.Context) {
	_m.Called(c)
}

// GetSheetAction provides a mock function with given fields: c
func (_m *ApiController) GetSheetAction(c *gin.Context) {
	_m.Called(c)
}

// SetCellAction provides a mock function with given fields: c
func (_m *ApiController) SetCellAction(c *gin.Context) {
	_m.Called(c)
}

type mockConstructorTestingTNewApiControllerInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewApiController creates a new instance of ApiController. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApiController(t mockConstructorTestingTNewApiControllerInterface) *ApiController {
	mock := &ApiController{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
