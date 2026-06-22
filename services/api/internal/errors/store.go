package errors

import "errors"

type StoreError error

var (
	ErrNotFound          StoreError = errors.New("not found")
	ClientNotInitialized StoreError = errors.New("deps missing")
)

func Custom(err error) StoreError {
	return err
}
