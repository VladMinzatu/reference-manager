package model

import (
	"errors"
	"fmt"
	"regexp"
)

type Id int64

func NewId(val int64) (Id, error) {
	if val <= 0 {
		return 0, errors.New("id must be positive")
	}
	return Id(val), nil
}

type Title string

const MaxTitleLength = 255

func NewTitle(val string) (Title, error) {
	if len(val) == 0 {
		return "", errors.New("title cannot be empty")
	}
	if len(val) > MaxTitleLength {
		return "", fmt.Errorf("title too long (max %d)", MaxTitleLength)
	}
	return Title(val), nil
}

type ISBN string

const MaxISBNLength = 50

func NewISBN(val string) (ISBN, error) {
	if len(val) == 0 {
		return "", errors.New("ISBN cannot be empty")
	}
	if len(val) > MaxISBNLength {
		return "", fmt.Errorf("ISBN too long (max %d)", MaxISBNLength)
	}
	return ISBN(val), nil
}

type URL string

var urlRegexp = regexp.MustCompile(`^(https?://)?[\w.-]+(\.[a-zA-Z]{2,})+.*$`)

func NewURL(val string) (URL, error) {
	if len(val) == 0 {
		return "", errors.New("URL cannot be empty")
	}

	if !urlRegexp.MatchString(val) {
		return "", errors.New("invalid URL format")
	}
	return URL(val), nil
}
