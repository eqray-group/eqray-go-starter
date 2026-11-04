<!-- 支付渠道页面：一行一种支付方式，凭据存库、后台可视化维护 -->
<template>
  <div class="pay-channel-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'pay:channel:create'" @click="showDialog('add')" v-ripple>
              新增渠道
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <PayChannelDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :channel-data="currentChannel"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchDeletePayChannel,
    fetchGetPayChannelList,
    fetchSetPayChannelStatus,
    fetchTestPayChannel
  } from '@/api/pay/channel'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import PayChannelDialog from './modules/pay-channel-dialog.vue'

  defineOptions({ name: 'PayChannel' })

  type ChannelItem = Api.Pay.ChannelItem
  type DialogType = 'add' | 'edit'

  // 渠道编码展示文案：后端配置的是编码，界面上给中文名
  const CODE_LABELS: Record<string, string> = {
    wx_native: '微信扫码',
    wx_h5: '微信 H5',
    wx_jsapi: '微信 JSAPI',
    wx_mini: '微信小程序',
    wx_app: '微信 APP',
    ali_pc: '支付宝电脑网站',
    ali_wap: '支付宝手机网站',
    ali_app: '支付宝 APP',
    ali_qr: '支付宝当面付'
  }

  // 支付方式展示文案
  const METHOD_LABELS: Record<string, string> = {
    native: '扫码',
    h5: 'H5',
    jsapi: 'JSAPI',
    mini: '小程序',
    app: 'APP',
    page: '电脑网站',
    wap: '手机网站',
    qr: '扫码'
  }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Pay.ChannelSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '渠道名称',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入渠道名称',
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '启用', value: 1 },
          { label: '停用', value: 0 }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentChannel = ref<ChannelItem | null>(null)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetPayChannelList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '渠道名称', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'code',
          label: '渠道编码',
          width: 180,
          formatter: (row: ChannelItem) =>
            h('div', { class: 'flex flex-col' }, [
              h(ElTag, { type: 'info' }, () => CODE_LABELS[row.code] || row.code),
              h('span', { class: 'mt-1 text-xs text-g-600' }, row.code)
            ])
        },
        {
          prop: 'payMethod',
          label: '支付方式',
          width: 110,
          formatter: (row: ChannelItem) => METHOD_LABELS[row.payMethod] || row.payMethod
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: ChannelItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '停用'
            )
        },
        { prop: 'sort', label: '排序', width: 80 },
        { prop: 'remark', label: '备注', minWidth: 160, showOverflowTooltip: true },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 300,
          fixed: 'right',
          formatter: (row: ChannelItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('pay:channel:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      icon: row.status === 1 ? 'ri:pause-circle-line' : 'ri:play-circle-line',
                      text: row.status === 1 ? '停用' : '启用',
                      onClick: () => handleToggleStatus(row)
                    }),
                    h(ArtButtonTable, {
                      type: 'action',
                      icon: 'ri:link',
                      text: '测通',
                      onClick: () => handleTest(row)
                    })
                  ]
                : []),
              ...(hasAuth('pay:channel:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('pay:channel:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Pay.ChannelSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: ChannelItem): void => {
    dialogType.value = type
    currentChannel.value = type === 'edit' ? (row ?? null) : null
    dialogVisible.value = true
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'edit') {
      refreshUpdate()
    } else {
      refreshCreate()
    }
  }

  // 启停渠道：停用后 C 端不可选、下单拒绝，但已发生的回调仍会处理
  const handleToggleStatus = async (row: ChannelItem): Promise<boolean | undefined> => {
    const enable = row.status !== 1
    try {
      await ElMessageBox.confirm(
        `确定要${enable ? '启用' : '停用'}渠道「${row.name}」吗？${
          enable ? '' : '停用后用户将无法选择该支付方式。'
        }`,
        `${enable ? '启用' : '停用'}渠道`,
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchSetPayChannelStatus(row.id, enable)
      ElMessage.success(`已${enable ? '启用' : '停用'}`)
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  // 连通性测试：用当前凭据发起一次查单，能区分「凭据错误」与「订单不存在」
  const handleTest = async (row: ChannelItem): Promise<void> => {
    try {
      const res = await fetchTestPayChannel(row.id)
      ElMessage.success(res.message || '测试通过')
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleDelete = async (row: ChannelItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除渠道「${row.name}」吗？删除后该支付方式立即不可用。`,
        '删除支付渠道',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
      )
    } catch {
      return
    }

    try {
      await fetchDeletePayChannel(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
