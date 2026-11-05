<!-- 基础设施 / 代码生成：选表 → 配置模块 → 预览 → 落盘 -->
<template>
  <div class="infra-generator-page art-full-height">
    <ElCard class="mb-4">
      <ElAlert type="warning" :closable="false" show-icon class="mb-4">
        <span class="text-xs">
          生成器按表结构产出 model / repository / service / request / controller + 前端 api 与页面；
          路由片段只预览、不落盘。已存在的文件默认跳过，勾选「覆盖」才会重写，请确认无手改逻辑。
        </span>
      </ElAlert>

      <ElForm label-position="top" class="gen-form">
        <div class="gen-grid">
          <ElFormItem label="数据表">
            <ElSelect
              v-model="form.table"
              filterable
              placeholder="请选择表"
              class="w-full"
              @change="handleTableChange"
            >
              <ElOption
                v-for="item in tables"
                :key="item.name"
                :label="`${item.name}${item.comment ? ' · ' + item.comment : ''}`"
                :value="item.name"
              />
            </ElSelect>
          </ElFormItem>

          <ElFormItem label="目标模块">
            <ElInput v-model="form.module" placeholder="如 demo / pay，对应 app/modules/{模块}" />
          </ElFormItem>

          <ElFormItem label="结构体名前缀（可选）">
            <ElInput v-model="form.pascal" placeholder="默认由表名推导，如 DemoOrder" />
          </ElFormItem>

          <ElFormItem label="覆盖已有文件">
            <ElSwitch v-model="form.overwrite" />
          </ElFormItem>
        </div>
      </ElForm>

      <div class="flex items-center gap-2">
        <ElButton
          type="primary"
          :loading="previewing"
          :disabled="!form.table"
          @click="handlePreview"
          v-ripple
        >
          预览代码
        </ElButton>
        <ElButton
          v-auth="'infra:generator:write'"
          :loading="generating"
          :disabled="!form.table || !form.module"
          @click="handleGenerate"
          v-ripple
        >
          生成落盘
        </ElButton>
        <span v-if="selectedTable" class="text-xs text-g-600">
          共 {{ columns.length }} 个字段，{{ writableCount }} 个文件会落盘
        </span>
      </div>
    </ElCard>

    <ElCard v-if="selectedTable" class="mb-4">
      <div class="panel-title">表结构 · {{ selectedTable.name }}</div>
      <ElTable :data="columns" size="small" border max-height="280">
        <ElTableColumn prop="name" label="字段" width="170" />
        <ElTableColumn prop="columnType" label="类型" width="150" />
        <ElTableColumn label="Go 类型" width="130">
          <template #default="{ row }">
            <span class="font-mono text-xs">{{ row.goType }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="表单" width="100">
          <template #default="{ row }">
            <ElTag size="small" type="info">{{ row.formType }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="主键" width="70">
          <template #default="{ row }">
            <ElTag v-if="row.primaryKey" size="small" type="warning">PK</ElTag>
            <span v-else class="text-g-600">-</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="comment" label="注释" min-width="160" show-overflow-tooltip />
      </ElTable>
    </ElCard>

    <ElCard v-if="files.length" class="art-table-card">
      <div class="flex items-center justify-between mb-3">
        <div class="panel-title !mb-0">生成结果（{{ files.length }} 个文件）</div>
        <span class="text-xs text-g-600">点击标签切换文件，可直接复制</span>
      </div>

      <ElTabs v-model="activeFile" class="gen-tabs">
        <ElTabPane v-for="file in files" :key="file.path" :name="file.path">
          <template #label>
            <span>{{ fileName(file.path) }}</span>
            <ElTag v-if="!file.writable" size="small" type="info" class="ml-1">仅预览</ElTag>
          </template>

          <div class="file-bar">
            <span class="file-path">{{ file.path }}</span>
            <ElButton size="small" @click="handleCopy(file)">复制</ElButton>
          </div>
          <pre class="code-block">{{ file.content }}</pre>
        </ElTabPane>
      </ElTabs>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElForm,
    ElFormItem,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElOption,
    ElSelect,
    ElSwitch,
    ElTable,
    ElTableColumn,
    ElTabPane,
    ElTabs,
    ElTag
  } from 'element-plus'
  import {
    fetchGenerateCode,
    fetchGetGeneratorColumns,
    fetchGetGeneratorTables,
    fetchPreviewGeneratorCode
  } from '@/api/infra/generator'

  defineOptions({ name: 'InfraGenerator' })

  type GeneratorFile = Api.Infra.GeneratorFile

  const tables = ref<Api.Infra.GeneratorTable[]>([])
  const columns = ref<Api.Infra.GeneratorColumn[]>([])
  const files = ref<GeneratorFile[]>([])
  const activeFile = ref('')
  const previewing = ref(false)
  const generating = ref(false)

  const form = reactive({
    table: '',
    module: '',
    pascal: '',
    overwrite: false
  })

  const selectedTable = computed(
    () => tables.value.find((item) => item.name === form.table) || null
  )

  const writableCount = computed(() => files.value.filter((file) => file.writable).length)

  const fileName = (path: string): string => path.split('/').slice(-1)[0] || path

  const loadTables = async (): Promise<void> => {
    try {
      tables.value = (await fetchGetGeneratorTables()) ?? []
    } catch {
      tables.value = []
    }
  }

  const handleTableChange = async (table: string): Promise<void> => {
    files.value = []
    columns.value = []

    // 默认模块取表名前缀（demo_order → demo），用户可改
    if (!form.module) {
      const prefix = table.split('_')[0]
      if (prefix) form.module = prefix
    }

    try {
      columns.value = (await fetchGetGeneratorColumns(table)) ?? []
    } catch {
      columns.value = []
    }
  }

  const handlePreview = async (): Promise<void> => {
    if (!form.module) {
      ElMessage.warning('请填写目标模块')
      return
    }

    previewing.value = true
    try {
      const res = await fetchPreviewGeneratorCode({
        table: form.table,
        module: form.module.trim(),
        pascal: form.pascal.trim() || undefined,
        overwrite: form.overwrite
      })
      files.value = res ?? []
      activeFile.value = files.value[0]?.path ?? ''
      if (!files.value.length) ElMessage.warning('没有生成内容')
    } catch {
      files.value = []
    } finally {
      previewing.value = false
    }
  }

  // 落盘前二次确认：会写真实文件，覆盖模式下风险更高
  const handleGenerate = async (): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        form.overwrite
          ? `将在 app/modules/${form.module} 与 ui/src 下生成文件，并覆盖同名文件。确认继续？`
          : `将在 app/modules/${form.module} 与 ui/src 下生成文件，已存在的文件会跳过。确认继续？`,
        '生成代码',
        { confirmButtonText: '生成', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    generating.value = true
    try {
      const res = await fetchGenerateCode({
        table: form.table,
        module: form.module.trim(),
        pascal: form.pascal.trim() || undefined,
        overwrite: form.overwrite
      })

      const written = res?.written?.length ?? 0
      const skipped = res?.skipped ?? []
      if (skipped.length > 0) {
        ElMessageBox.alert(
          `已写入 ${written} 个文件；以下 ${skipped.length} 个已存在被跳过：\n${skipped.join('\n')}`,
          '生成完成',
          { confirmButtonText: '知道了' }
        )
      } else {
        ElMessage.success(`生成完成，写入 ${written} 个文件`)
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      generating.value = false
    }
  }

  // 复制文件内容到剪贴板
  const copyText = async (text?: string): Promise<void> => {
    if (!text) return
    try {
      await navigator.clipboard.writeText(text)
      ElMessage.success('已复制')
    } catch {
      ElMessage.warning('复制失败，请手动选中复制')
    }
  }

  const handleCopy = (file: GeneratorFile): void => {
    void copyText(file.content)
  }

  onMounted(loadTables)
</script>

<style lang="scss" scoped>
  .gen-form {
    :deep(.el-form-item) {
      margin-bottom: 12px;
    }

    :deep(.el-form-item__label) {
      padding-bottom: 4px;
      font-weight: 500;
    }
  }

  .gen-grid {
    display: grid;
    gap: 0 16px;
    grid-template-columns: repeat(4, minmax(0, 1fr));

    @media (max-width: 1280px) {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    @media (max-width: 640px) {
      grid-template-columns: 1fr;
    }
  }

  .panel-title {
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .file-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 0 8px;
  }

  .file-path {
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .code-block {
    max-height: 520px;
    padding: 14px 16px;
    margin: 0;
    overflow: auto;
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
    line-height: 1.7;
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
  }

  .gen-tabs {
    :deep(.el-tabs__header) {
      margin-bottom: 8px;
    }

    :deep(.el-tabs__item) {
      height: 36px;
      font-size: 13px;
    }
  }
</style>
