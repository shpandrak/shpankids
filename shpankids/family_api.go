package shpankids

import "context"

type FamilyDto struct {
	Id   string
	Name string
}

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string) error
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)
}
