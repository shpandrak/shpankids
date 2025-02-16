package shpankids

import (
	"context"
	"time"
)

type Session struct {
	FamilyId string
	Location *time.Location
}

type SessionManager interface {
	Get(ctx context.Context, userId string) (*Session, error)
	Set(ctx context.Context, userId string, session Session) error
}
