import request from '@/utils/http'

// 渠道类型与配置字段定义（前端据此动态渲染表单）
export function fetchGetNoticeTypes() {
  return request.get<Api.Infra.NoticeTypeSchema[]>({
    url: '/api/v1/infra/notice/types'
  })
}

// 分页查询通知渠道
export function fetchGetNoticeChannels(params: Api.Infra.NoticeSearchParams) {
  return request.get<Api.Infra.NoticeChannelList>({
    url: '/api/v1/infra/notice-channels',
    params
  })
}

// 新增渠道
export function fetchCreateNoticeChannel(data: Api.Infra.NoticeChannelSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/notice-channels',
    params: data
  })
}

// 编辑渠道
export function fetchUpdateNoticeChannel(id: string, data: Api.Infra.NoticeChannelSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/notice-channels/${id}`,
    params: data
  })
}

// 删除渠道
export function fetchDeleteNoticeChannel(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/notice-channels/${id}`
  })
}

// 测试发送（同步返回发送记录）
export function fetchTestNoticeChannel(id: string, target?: string) {
  return request.post<Api.Infra.NoticeLogItem>({
    url: `/api/v1/infra/notice-channels/${id}/test`,
    params: { target },
    timeout: 60000
  })
}

// 分页查询模板
export function fetchGetNoticeTemplates(params: Api.Infra.NoticeSearchParams) {
  return request.get<Api.Infra.NoticeTemplateList>({
    url: '/api/v1/infra/notice-templates',
    params
  })
}

// 新增模板
export function fetchCreateNoticeTemplate(data: Api.Infra.NoticeTemplateSaveParams) {
  return request.post<{ id: string }>({
    url: '/api/v1/infra/notice-templates',
    params: data
  })
}

// 编辑模板
export function fetchUpdateNoticeTemplate(id: string, data: Api.Infra.NoticeTemplateSaveParams) {
  return request.put<null>({
    url: `/api/v1/infra/notice-templates/${id}`,
    params: data
  })
}

// 删除模板
export function fetchDeleteNoticeTemplate(id: string) {
  return request.del<null>({
    url: `/api/v1/infra/notice-templates/${id}`
  })
}

// 分页查询发送记录
export function fetchGetNoticeLogs(params: Api.Infra.NoticeSearchParams) {
  return request.get<Api.Infra.NoticeLogList>({
    url: '/api/v1/infra/notice-logs',
    params
  })
}

// 通用发送
export function fetchSendNotice(data: Api.Infra.NoticeSendParams) {
  return request.post<Api.Infra.NoticeLogItem>({
    url: '/api/v1/infra/notices/send',
    params: data,
    timeout: 60000
  })
}
