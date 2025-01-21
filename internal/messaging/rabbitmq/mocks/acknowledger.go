package mocks

import (
	"github.com/stretchr/testify/mock"
)

type Acknowledger struct {
	mock.Mock

	AckedTags map[uint64]bool
}

func NewAcknowledger(t interface {
	mock.TestingT
	Cleanup(func())
}) *Acknowledger {
	mock := &Acknowledger{AckedTags: make(map[uint64]bool)}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

func (a *Acknowledger) Ack(tag uint64, multiple bool) error {
	a.AckedTags[tag] = true
	return a.Called(tag, multiple).Error(0)
}

func (a *Acknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return a.Called(tag, multiple, requeue).Error(0)
}

func (a *Acknowledger) Reject(tag uint64, requeue bool) error {
	return a.Called(tag, requeue).Error(0)
}
