package model

import "errors"

var ErrConcurrentCategoryUpdate = errors.New("concurrent update error on category")

type Category struct {
	Id         Id
	Name       Title
	References []Reference
	Version    Version
}
