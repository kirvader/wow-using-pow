package storage

import (
	"context"
	"time"
)

var _ StorageSet = &StorageSetMock{}

type StorageSetMock struct {
	InsertClientTokenFunc func(context.Context, string, string, time.Duration) error
	GetClientTokenFunc    func(context.Context, string) (string, error)
	DeleteFunc            func(context.Context, string) error
	CloseFunc             func(context.Context) error
}

func (storageMock *StorageSetMock) InsertClientToken(ctx context.Context, key, value string, dur time.Duration) error {
	return storageMock.InsertClientTokenFunc(ctx, key, value, dur)
}
func (storageMock *StorageSetMock) GetClientToken(ctx context.Context, key string) (string, error) {
	return storageMock.GetClientTokenFunc(ctx, key)
}
func (storageMock *StorageSetMock) Delete(ctx context.Context, key string) error {
	return storageMock.DeleteFunc(ctx, key)
}

func (storageMock *StorageSetMock) Close(ctx context.Context) error {
	return storageMock.CloseFunc(ctx)
}
