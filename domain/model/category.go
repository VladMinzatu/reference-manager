package model

import "errors"

var ErrConcurrentCategoryUpdate = errors.New("concurrent update error on category")

type Category struct {
	Id         Id
	Name       Title
	References []Reference
	Version    Version
}

// CategoryRef is a lightweight view of a category (id + title only) that *might* come in handy, wink, wink
type CategoryRef struct {
	Id   Id
	Name Title
}
