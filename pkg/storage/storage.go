package storage

import (
	"context"
	"time"
)

type StorageSet interface {
	Add(context.Context, string, time.Duration) error
	Exists(context.Context, string) (bool, error)
	Delete(context.Context, string) error
	Close(context.Context) error
}
