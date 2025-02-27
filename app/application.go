package app

import (
	"fmt"
	"shpankids/domain/assignment"
	"shpankids/domain/family"
	"shpankids/domain/session"
	"shpankids/domain/user"
	"shpankids/infra/database/kvstore"
	"shpankids/webserver"
	"shpankids/webserver/auth"
)

func Start(kvs kvstore.RawJsonStore) error {
	userManager := user.NewUserManager(kvs)
	familyManager := family.NewFamilyManager(kvs, auth.GetUserInfo)
	sessionManager := session.NewSessionManager(kvs)
	assignmentManager := assignment.NewAssignmentManager(kvs, auth.GetUserInfo, familyManager, sessionManager)

	err := appBootstrap(userManager, familyManager, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to bootstrap app: %v", err)
	}
	return webserver.Start(assignmentManager, userManager, familyManager, sessionManager)
}
