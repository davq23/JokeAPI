package repositories

import "errors"

type FetchDirection int8

const FetchNext FetchDirection = 1
const FetchBack FetchDirection = -1

var ErrInvalidOffset error = errors.New("invalid offset ID")