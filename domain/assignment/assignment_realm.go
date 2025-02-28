package assignment

import "shpankids/shpankids"

type Realm interface {
	GetUserRole() shpankids.Role
	GetRealmUserIds() []string
}
