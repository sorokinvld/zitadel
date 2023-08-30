package mock

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/zitadel/zitadel/internal/eventstore/repository"
)

func NewRepo(t *testing.T) *MockRepository {
	return NewMockRepository(gomock.NewController(t))
}

func (m *MockRepository) ExpectFilterNoEventsNoError() *MockRepository {
	m.EXPECT().Filter(gomock.Any(), gomock.Any()).Return(nil, nil)
	return m
}

func (m *MockRepository) ExpectFilterEvents(events ...*repository.Event) *MockRepository {
	m.EXPECT().Filter(gomock.Any(), gomock.Any()).Return(events, nil)
	return m
}

func (m *MockRepository) ExpectFilterEventsError(err error) *MockRepository {
	m.EXPECT().Filter(gomock.Any(), gomock.Any()).Return(nil, err)
	return m
}

func (m *MockRepository) ExpectInstanceIDs(hasFilters []*repository.Filter, instanceIDs ...string) *MockRepository {
	matcher := gomock.Any()
	if len(hasFilters) > 0 {
		matcher = &filterQueryMatcher{Filters: [][]*repository.Filter{hasFilters}}
	}
	m.EXPECT().InstanceIDs(gomock.Any(), matcher).Return(instanceIDs, nil)
	return m
}

func (m *MockRepository) ExpectInstanceIDsError(err error) *MockRepository {
	m.EXPECT().InstanceIDs(gomock.Any(), gomock.Any()).Return(nil, err)
	return m
}

func (m *MockRepository) ExpectPush(expectedEvents []*repository.Event, expectedUniqueConstraints ...*repository.UniqueConstraint) *MockRepository {
	m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, events []*repository.Event, uniqueConstraints ...*repository.UniqueConstraint) error {
			assert.Equal(m.ctrl.T, expectedEvents, events)
			if expectedUniqueConstraints == nil {
				expectedUniqueConstraints = []*repository.UniqueConstraint{}
			}
			assert.Equal(m.ctrl.T, expectedUniqueConstraints, uniqueConstraints)
			return nil
		},
	)
	return m
}

func (m *MockRepository) ExpectPushFailed(err error, expectedEvents []*repository.Event, expectedUniqueConstraints ...*repository.UniqueConstraint) *MockRepository {
	m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, events []*repository.Event, uniqueConstraints ...*repository.UniqueConstraint) error {
			assert.Equal(m.ctrl.T, expectedEvents, events)
			if expectedUniqueConstraints == nil {
				expectedUniqueConstraints = []*repository.UniqueConstraint{}
			}
			assert.Equal(m.ctrl.T, expectedUniqueConstraints, uniqueConstraints)
			return err
		},
	)
	return m
}

func (m *MockRepository) ExpectRandomPush(expectedEvents []*repository.Event, expectedUniqueConstraints ...*repository.UniqueConstraint) *MockRepository {
	m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, events []*repository.Event, uniqueConstraints ...*repository.UniqueConstraint) error {
			assert.Len(m.ctrl.T, events, len(expectedEvents))
			assert.Len(m.ctrl.T, expectedUniqueConstraints, len(uniqueConstraints))
			return nil
		},
	)
	return m
}

func (m *MockRepository) ExpectRandomPushFailed(err error, expectedEvents []*repository.Event, expectedUniqueConstraints ...*repository.UniqueConstraint) *MockRepository {
	m.EXPECT().Push(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, events []*repository.Event, uniqueConstraints ...*repository.UniqueConstraint) error {
			assert.Len(m.ctrl.T, events, len(expectedEvents))
			assert.Len(m.ctrl.T, expectedUniqueConstraints, len(uniqueConstraints))
			return err
		},
	)
	return m
}
