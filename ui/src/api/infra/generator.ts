import request from '@/utils/http'

// 可生成的业务表
export function fetchGetGeneratorTables() {
  return request.get<Api.Infra.GeneratorTable[]>({
    url: '/api/v1/infra/generator/tables'
  })
}

// 表字段结构
export function fetchGetGeneratorColumns(table: string) {
  return request.get<Api.Infra.GeneratorColumn[]>({
    url: `/api/v1/infra/generator/tables/${table}/columns`
  })
}

// 预览生成内容（不落盘）
export function fetchPreviewGeneratorCode(data: Api.Infra.GeneratorParams) {
  return request.post<Api.Infra.GeneratorFile[]>({
    url: '/api/v1/infra/generator/preview',
    params: data
  })
}

// 生成并写入文件
export function fetchGenerateCode(data: Api.Infra.GeneratorParams) {
  return request.post<Api.Infra.GeneratorResult>({
    url: '/api/v1/infra/generator/generate',
    params: data
  })
}
