package myid

import "github.com/segmentio/ksuid"

func New() string {
	return ksuid.New().String()
}
