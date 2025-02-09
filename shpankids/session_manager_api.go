package shpankids

import (
	"context"
)

type Session struct {
	FamilyId string
}

type SessionManager interface {
	Get(ctx context.Context, userId string) (*Session, error)
	Set(ctx context.Context, userId string, session Session) error
}
