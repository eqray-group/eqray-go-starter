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
		&migrations.M20261017000009SeedMenuIconAndDashboard{},
		&migrations.M20261017000010DropDeletedKeyColumns{},
		&migrations.M20261019000001CreateAuthLoginLogTable{},
		&migrations.M20261019000002CreateSystemDictTypeTable{},
		&migrations.M20261019000003CreateSystemDictDataTable{},
		&migrations.M20261019000004SeedDictData{},
		&migrations.M20261019000005SeedDictAndLoginLogMenu{},
		&migrations.M20261020000001CreateInfraFileTable{},
		&migrations.M20261020000002CreateSystemOperateLogTable{},
		&migrations.M20261020000003SeedOperateLogMenu{},
		&migrations.M20261020000004CreateInfraConfigTable{},
		&migrations.M20261020000005SeedInfraConfigMenu{},
		&migrations.M20261020000006SeedInfraFileMenu{},
		&migrations.M20261020000007CreateInfraFileConfigTable{},
		&migrations.M20261020000008CreateInfraApiAccessLogTable{},
		&migrations.M20261020000009CreateInfraApiErrorLogTable{},
		&migrations.M20261020000012SeedInfraOpsMenus{},
		&migrations.M20261020000013SeedInfraFileConfig{},
	}
}
