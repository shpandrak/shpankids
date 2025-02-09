package auth

import (
	"context"
	"fmt"
)

func GetUserInfo(ctx context.Context) (*string, error) {
	value := ctx.Value("x-shpankids-user")
	if value == nil {
		return nil, fmt.Errorf("user not authenticated")
	}
	email, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("user not authenticated (invalid)")
	}
	return &email, nil
}

func EnrichContext(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, "x-shpankids-user", email)
}
