package service

import "context"

// New creates the default AWS-backed service.
func New(ctx context.Context) (Service, error) {
	return NewAWS(ctx)
}
