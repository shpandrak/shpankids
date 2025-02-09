package shpankids

import "context"

type UserSessionManager func(ctx context.Context) (*string, error)
