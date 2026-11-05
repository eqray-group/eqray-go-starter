package model

import (
	"eqray-go-starter/app/modules/base/types"
)

// 定时任务表名常量。
const (
	// InfraJobTable 定时任务表
	InfraJobTable = "infra_job"
	// InfraJobLogTable 定时任务执行日志表
	InfraJobLogTable = "infra_job_log"
)

// 任务状态。
const (
	// InfraJobStatusDisabled 停用
	InfraJobStatusDisabled int8 = 0
	// InfraJobStatusEnabled 启用
	InfraJobStatusEnabled int8 = 1
)

// 执行结果。
const (
	// InfraJobResultFailed 失败
	InfraJobResultFailed int8 = 0
	// InfraJobResultSuccess 成功
	InfraJobResultSuccess int8 = 1
)

// 触发方式。
const (
	// InfraJobTriggerCron 调度触发
	InfraJobTriggerCron = "cron"
	// InfraJobTriggerManual 手动触发
	InfraJobTriggerManual = "manual"
)

// InfraJobModel 对应 infra_job 表：一条记录代表一个可调度任务。
type InfraJobModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// Name 任务名称
	Name string `db:"name" json:"name"`
	// Handler 处理器标识，对应代码里注册的 Key
	Handler string `db:"handler" json:"handler"`
	// Cron cron 表达式（5 段）
	Cron string `db:"cron" json:"cron"`
	// Status 状态：0 停用 / 1 启用
	Status int8 `db:"status" json:"status"`
	// Remark 备注
	Remark string `db:"remark" json:"remark"`
	// NextRunAt 下次执行时间
	NextRunAt *types.DateTime `db:"next_run_at" json:"nextRunAt"`
	// LastRunAt 上次执行时间
	LastRunAt *types.DateTime `db:"last_run_at" json:"lastRunAt"`
	// LastStatus 上次结果：0 失败 / 1 成功
	LastStatus int8 `db:"last_status" json:"lastStatus"`
	// LastDuration 上次耗时（毫秒）
	LastDuration int `db:"last_duration" json:"lastDuration"`
	// LastOutput 上次输出摘要
	LastOutput string `db:"last_output" json:"lastOutput"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
	// Updater 更新人
	Updater string `db:"updater" json:"updater"`
	// UpdatedAt 更新时间
	UpdatedAt types.DateTime `db:"updated_at" json:"updatedAt"`
}

// TableName 返回表名。
func (m *InfraJobModel) TableName() string {
	return InfraJobTable
}

// InfraJobLogModel 对应 infra_job_log 表：一次执行一行，只追加。
type InfraJobLogModel struct {
	// ID 主键
	ID types.ID `db:"id" json:"id"`
	// JobID 任务 ID
	JobID types.ID `db:"job_id" json:"jobId"`
	// JobName 任务名称（冗余，删任务后仍可查）
	JobName string `db:"job_name" json:"jobName"`
	// Handler 处理器标识
	Handler string `db:"handler" json:"handler"`
	// Trigger 触发方式：cron / manual
	Trigger string `db:"trigger" json:"trigger"`
	// Status 结果：0 失败 / 1 成功
	Status int8 `db:"status" json:"status"`
	// StartedAt 开始时间
	StartedAt types.DateTime `db:"started_at" json:"startedAt"`
	// FinishedAt 结束时间
	FinishedAt types.DateTime `db:"finished_at" json:"finishedAt"`
	// Duration 耗时（毫秒）
	Duration int `db:"duration" json:"duration"`
	// Output 执行输出
	Output string `db:"output" json:"output"`
	// Error 失败原因
	Error string `db:"error" json:"error"`
	// CreatedAt 创建时间
	CreatedAt types.DateTime `db:"created_at" json:"createdAt"`
}

// TableName 返回表名。
func (m *InfraJobLogModel) TableName() string {
	return InfraJobLogTable
}
