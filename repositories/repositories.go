package repositories

import "errors"

type FetchDirection int8

const FetchNext FetchDirection = 1
const FetchBack FetchDirection = -1

var ErrInvalidOffset error = errors.New("invalid offset ID")
var ErrUnknownID error = errors.New("unknown ID")
var ErrUnknownEmail error = errors.New("unknown email")
