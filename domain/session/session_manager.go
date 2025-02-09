package session

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
)

type manager struct {
	repository sessionRepository
}

func NewSessionManager(kvs kvstore.RawJsonStore) shpankids.SessionManager {
	return &manager{
		repository: newSessionRepository(kvs),
	}
}

func (m *manager) Get(ctx context.Context, email string) (*shpankids.Session, error) {
	dbS, err := m.repository.Get(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapSession(dbS), nil

}

func mapSession(s dbSession) *shpankids.Session {
	return &shpankids.Session{
		FamilyId: s.FamilyId,
	}
}

func (m *manager) Set(ctx context.Context, userId string, session shpankids.Session) error {
	return m.repository.Set(ctx, userId, dbSession{
		FamilyId: session.FamilyId,
	})

}
