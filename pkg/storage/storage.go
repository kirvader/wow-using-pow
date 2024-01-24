package storage

import (
	"context"
	"time"
)

type StorageSet interface {
	InsertClientToken(context.Context, string, string, time.Duration) error
	GetClientToken(context.Context, string) (string, error)
	Delete(context.Context, string) error
	Close(context.Context) error
}
