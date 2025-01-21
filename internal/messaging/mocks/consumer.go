// Code generated by mockery v2.33.3. DO NOT EDIT.

package mocks

import (
	amqp091 "github.com/rabbitmq/amqp091-go"

	mock "github.com/stretchr/testify/mock"
)

// Consumer is an autogenerated mock type for the Consumer type
type Consumer struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Consumer) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Consume provides a mock function with given fields: queue
func (_m *Consumer) Consume(queue string) (<-chan amqp091.Delivery, error) {
	ret := _m.Called(queue)

	var r0 <-chan amqp091.Delivery
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (<-chan amqp091.Delivery, error)); ok {
		return rf(queue)
	}
	if rf, ok := ret.Get(0).(func(string) <-chan amqp091.Delivery); ok {
		r0 = rf(queue)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan amqp091.Delivery)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(queue)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewConsumer creates a new instance of Consumer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewConsumer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Consumer {
	mock := &Consumer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
