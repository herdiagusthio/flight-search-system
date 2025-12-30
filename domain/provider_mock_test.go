package domain

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestMockFlightProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockFlightProvider(ctrl)

	mock.EXPECT().Name().Return("test")
	mock.Name()

	mock.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, nil)
	mock.Search(context.Background(), SearchCriteria{})
}

func TestMockProviderRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockProviderRegistry(ctrl)

	mock.EXPECT().Register(nil)
	mock.Register(nil)

	mock.EXPECT().Get("test").Return(nil)
	mock.Get("test")

	mock.EXPECT().GetAll().Return(nil)
	mock.GetAll()

	mock.EXPECT().Names().Return(nil)
	mock.Names()
}
