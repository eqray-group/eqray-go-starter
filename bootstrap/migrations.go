package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"eqray-go-starter/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20261017000001CreateSystemUsersTable{},
		&migrations.M20261017000002CreateSystemMenuTable{},
		&migrations.M20261017000003CreateSystemRoleTable{},
		&migrations.M20261017000004CreateSystemUserRoleTable{},
		&migrations.M20261017000005CreateSystemRoleMenuTable{},
		&migrations.M20261017000006SeedSystemMenu{},
		&migrations.M20261017000007RenameSuperAdminRoleCode{},
	}
}
