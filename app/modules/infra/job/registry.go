// Package job 提供定时任务的处理器注册表。
// 库里只存「处理器标识 + cron 表达式」，具体逻辑在 bootstrap/jobs.go 里注册，
// 这样改频率 / 启停可由运营自助完成，改逻辑仍走代码评审。
package job

import (
	"context"
	"sort"
	"sync"
)

// Handler 是任务处理器：返回一段执行摘要（会截断后写入日志与列表），error 表示失败。
type Handler func(ctx context.Context) (string, error)

// Definition 描述一个可被调度器识别的任务。
type Definition struct {
	// Key 处理器标识，落库 infra_job.handler，必须全局唯一
	Key string
	// Name 默认任务名称，新建任务时作为默认值
	Name string
	// Description 说明，展示在后台「新增任务」的处理器选择里
	Description string
	// DefaultCron 默认 cron 表达式（5 段）
	DefaultCron string
	// Handler 实际执行逻辑
	Handler Handler
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Definition)
)

// Register 注册一个任务处理器；重复注册同一 Key 会覆盖，便于测试时替换。
// 入参 def 为任务定义；无返回值。
func Register(def Definition) {
	mu.Lock()
	defer mu.Unlock()

	registry[def.Key] = def
}

// Get 按处理器标识查询定义。
// 入参 key 为处理器标识；返回定义与是否存在。
func Get(key string) (Definition, bool) {
	mu.RLock()
	defer mu.RUnlock()

	def, ok := registry[key]

	return def, ok
}

// All 返回全部已注册处理器，按 Key 升序，供后台下拉选择。
// 入参无；返回定义列表。
func All() []Definition {
	mu.RLock()
	defer mu.RUnlock()

	list := make([]Definition, 0, len(registry))
	for _, def := range registry {
		list = append(list, def)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Key < list[j].Key })

	return list
}
