package service

import (
	"context"
	"runtime"
	"time"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/infra/model"
)

// monitorStartedAt 是进程启动时间，用于计算运行时长。
var monitorStartedAt = time.Now()

// MonitorAppStat 是应用进程信息。
type MonitorAppStat struct {
	// GoVersion Go 版本
	GoVersion string `json:"goVersion"`
	// UptimeSeconds 已运行秒数
	UptimeSeconds int64 `json:"uptimeSeconds"`
	// Goroutines 当前协程数
	Goroutines int `json:"goroutines"`
	// CPUCores 可用 CPU 核数
	CPUCores int `json:"cpuCores"`
}

// MonitorMemoryStat 是进程内存占用（单位 MB）。
type MonitorMemoryStat struct {
	// AllocMB 当前堆上活跃对象占用
	AllocMB float64 `json:"allocMB"`
	// HeapInuseMB 已向系统申请、正在使用的堆内存
	HeapInuseMB float64 `json:"heapInuseMB"`
	// SysMB 进程向系统申请的总内存
	SysMB float64 `json:"sysMB"`
	// NumGC GC 次数
	NumGC uint32 `json:"numGC"`
	// HeapObjects 堆对象数
	HeapObjects uint64 `json:"heapObjects"`
}

// MonitorHealthStat 是外部依赖的连通性探测结果。
type MonitorHealthStat struct {
	// OK 是否可用
	OK bool `json:"ok"`
	// LatencyMs 探测耗时（毫秒）
	LatencyMs int64 `json:"latencyMs"`
	// Message 失败原因，正常时为空
	Message string `json:"message"`
}

// MonitorRequestStat 是近 24 小时的请求统计。
type MonitorRequestStat struct {
	// Total 请求总数
	Total int64 `json:"total"`
	// Errors 失败请求数（result_code >= 400）
	Errors int64 `json:"errors"`
	// Slow 慢请求数（耗时 >= 1000ms）
	Slow int64 `json:"slow"`
	// AvgDurationMs 平均耗时
	AvgDurationMs float64 `json:"avgDurationMs"`
}

// MonitorErrorItem 是最近一条错误日志的摘要。
type MonitorErrorItem struct {
	// ID 错误日志 ID
	ID string `json:"id"`
	// Name 错误名称
	Name string `json:"name"`
	// RequestURL 出错接口
	RequestURL string `json:"requestUrl"`
	// Message 错误信息
	Message string `json:"message"`
	// CreatedAt 发生时间
	CreatedAt string `json:"createdAt"`
}

// MonitorOverview 是系统监控页的全部数据。
type MonitorOverview struct {
	// App 进程信息
	App MonitorAppStat `json:"app"`
	// Memory 内存占用
	Memory MonitorMemoryStat `json:"memory"`
	// DB 数据库探测
	DB MonitorHealthStat `json:"db"`
	// Cache 缓存探测
	Cache MonitorHealthStat `json:"cache"`
	// Request 近 24 小时请求统计
	Request MonitorRequestStat `json:"request"`
	// RecentErrors 最近 10 条错误日志
	RecentErrors []MonitorErrorItem `json:"recentErrors"`
	// CollectedAt 采集时间
	CollectedAt string `json:"collectedAt"`
}

// MonitorService 汇总进程、依赖与业务日志指标，供后台「系统监控」页展示。
// 不做长时采样，只读当前值，避免监控本身成为负担。
type MonitorService struct{}

// NewMonitorService 构造监控服务。
func NewMonitorService() *MonitorService {
	return &MonitorService{}
}

// Overview 采集一次监控数据。
// 入参 ctx 为请求上下文；返回汇总结果（单项探测失败不影响整体）。
func (s *MonitorService) Overview(ctx context.Context) *MonitorOverview {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	overview := &MonitorOverview{
		App: MonitorAppStat{
			GoVersion:     runtime.Version(),
			UptimeSeconds: int64(time.Since(monitorStartedAt).Seconds()),
			Goroutines:    runtime.NumGoroutine(),
			CPUCores:      runtime.NumCPU(),
		},
		Memory: MonitorMemoryStat{
			AllocMB:     toMB(mem.Alloc),
			HeapInuseMB: toMB(mem.HeapInuse),
			SysMB:       toMB(mem.Sys),
			NumGC:       mem.NumGC,
			HeapObjects: mem.HeapObjects,
		},
		DB:           s.checkDB(ctx),
		Cache:        s.checkCache(),
		Request:      s.requestStat(ctx),
		RecentErrors: s.recentErrors(ctx),
		CollectedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	return overview
}

// checkDB 通过一次轻量计数探测数据库连通性与延迟。
// 入参 ctx 为请求上下文；返回探测结果。
func (s *MonitorService) checkDB(ctx context.Context) MonitorHealthStat {
	started := time.Now()

	if _, err := facades.DB().WithContext(ctx).Table("system_menu").Count(); err != nil {
		return MonitorHealthStat{
			OK:        false,
			LatencyMs: time.Since(started).Milliseconds(),
			Message:   err.Error(),
		}
	}

	return MonitorHealthStat{OK: true, LatencyMs: time.Since(started).Milliseconds()}
}

// checkCache 通过写入 + 读取一次探测缓存连通性与延迟。
// 入参无；返回探测结果。
func (s *MonitorService) checkCache() MonitorHealthStat {
	const key = "infra:monitor:ping"

	started := time.Now()
	if err := facades.Cache().Put(key, time.Now().Unix(), 10*time.Second); err != nil {
		return MonitorHealthStat{
			OK:        false,
			LatencyMs: time.Since(started).Milliseconds(),
			Message:   err.Error(),
		}
	}
	if facades.Cache().Get(key) == nil {
		return MonitorHealthStat{
			OK:        false,
			LatencyMs: time.Since(started).Milliseconds(),
			Message:   "写入成功但读取为空，请检查缓存驱动",
		}
	}

	return MonitorHealthStat{OK: true, LatencyMs: time.Since(started).Milliseconds()}
}

// requestStat 统计近 24 小时的请求量、失败数与平均耗时。
// 入参 ctx 为请求上下文；返回统计结果（表不存在时返回空值）。
func (s *MonitorService) requestStat(ctx context.Context) MonitorRequestStat {
	if !facades.Schema().HasTable(model.InfraApiAccessLogTable) {
		return MonitorRequestStat{}
	}

	since := time.Now().Add(-24 * time.Hour)
	row := struct {
		Total int64   `db:"total"`
		Avg   float64 `db:"avg_duration"`
	}{}
	_ = facades.DB().WithContext(ctx).Table(model.InfraApiAccessLogTable).
		Select("COUNT(*) AS total, IFNULL(AVG(duration), 0) AS avg_duration").
		Where("created_at", ">=", since).
		First(&row)

	errors, _ := facades.DB().WithContext(ctx).Table(model.InfraApiAccessLogTable).
		Where("created_at", ">=", since).
		Where("result_code", ">=", 400).
		Count()

	slow, _ := facades.DB().WithContext(ctx).Table(model.InfraApiAccessLogTable).
		Where("created_at", ">=", since).
		Where("duration", ">=", 1000).
		Count()

	return MonitorRequestStat{
		Total:         row.Total,
		Errors:        errors,
		Slow:          slow,
		AvgDurationMs: float64(int64(row.Avg*100)) / 100,
	}
}

// recentErrors 取最近 10 条接口错误日志。
// 入参 ctx 为请求上下文；返回错误摘要列表。
func (s *MonitorService) recentErrors(ctx context.Context) []MonitorErrorItem {
	items := make([]MonitorErrorItem, 0, 10)
	if !facades.Schema().HasTable(model.InfraApiErrorLogTable) {
		return items
	}

	rows := make([]model.InfraApiErrorLogModel, 0, 10)
	if err := facades.DB().WithContext(ctx).Table(model.InfraApiErrorLogTable).
		OrderBy("created_at", "desc").
		Limit(10).
		Get(&rows); err != nil {
		return items
	}

	for _, row := range rows {
		items = append(items, MonitorErrorItem{
			ID:         row.ID.String(),
			Name:       row.Name,
			RequestURL: row.RequestMethod + " " + row.RequestURL,
			Message:    row.Message,
			CreatedAt:  row.CreatedAt.String(),
		})
	}

	return items
}

// toMB 把字节数换算成 MB，保留两位小数。
// 入参 bytes 为字节数；返回 MB 值。
func toMB(bytes uint64) float64 {
	return float64(int64(float64(bytes)/1024/1024*100)) / 100
}
