package session

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const sessionValueKey = "shpanCtx-userSession"

type manager struct {
	repository sessionRepository
}

func NewSessionManager(kvs kvstore.RawJsonStore) shpankids.SessionManager {
	return &manager{
		repository: newSessionRepository(kvs),
	}
}

func (m *manager) Get(ctx context.Context, email string) (*shpankids.Session, error) {

	// Check first if already in ctx
	if s, ok := ctx.Value(sessionValueKey).(*shpankids.Session); ok {
		return s, nil
	}

	dbS, err := m.repository.Get(ctx, email)
	if err != nil {
		return nil, err
	}

	// Save in ctx
	ctx = context.WithValue(ctx, sessionValueKey, dbS)

	return mapSession(dbS)

}

func mapSession(s dbSession) (*shpankids.Session, error) {
	l, err := time.LoadLocation(s.TimeZone)
	if err != nil {
		return nil, err
	}
	return &shpankids.Session{
		FamilyId: s.FamilyId,
		Location: l,
	}, nil
}

func (m *manager) Set(ctx context.Context, userId string, session shpankids.Session) error {
	return m.repository.Set(ctx, userId, dbSession{
		FamilyId: session.FamilyId,
		TimeZone: session.Location.String(),
	})

}
