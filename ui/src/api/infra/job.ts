import request from '@/utils/http'

// 分页查询定时任务
export function fetchGetJobList(params: Api.Infra.JobSearchParams) {
  return request.get<Api.Infra.JobList>({
    url: '/api/v1/infra/jobs',
    params
  })
}

// 可选的处理器清单（代码注册）
export function fetchGetJobHandlers() {
  return request.get<Api.Infra.JobHandlerItem[]>({
    url: '/api/v1/infra/jobs/handlers'
  })
}

// 新增任务
export function fetchCreateJob(data: Api.Infra.JobSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/jobs',
    params: data
  })
}

// 编辑任务（处理器标识不可改）
export function fetchUpdateJob(id: string, data: Api.Infra.JobSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/jobs/${id}`,
    params: data
  })
}

// 启停任务
export function fetchUpdateJobStatus(id: string, status: number) {
  return request.put<null>({
    url: `/api/v1/infra/jobs/${id}/status`,
    params: { status }
  })
}

// 删除任务
export function fetchDeleteJob(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/jobs/${id}`
  })
}

// 手动执行一次（同步，返回执行日志）
export function fetchRunJob(id: string) {
  return request.post<Api.Infra.JobLogItem>({
    url: `/api/v1/infra/jobs/${id}/run`,
    timeout: 120000
  })
}

// 分页查询执行日志
export function fetchGetJobLogs(params: Api.Infra.JobLogSearchParams) {
  return request.get<Api.Infra.JobLogList>({
    url: '/api/v1/infra/jobs/logs',
    params
  })
}
