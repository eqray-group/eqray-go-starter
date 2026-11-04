<!-- 支付渠道页面：卡片式渠道矩阵，一行一个渠道配置，不再用表格 -->
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
      <div class="channel-toolbar">
        <ElRadioGroup v-model="activeVendor" size="default">
          <ElRadioButton value="">全部（{{ data.length }}）</ElRadioButton>
          <ElRadioButton value="wx">微信支付</ElRadioButton>
          <ElRadioButton value="alipay">支付宝</ElRadioButton>
        </ElRadioGroup>

        <ElSpace>
          <ElButton circle :loading="loading" title="刷新" @click="refreshData">
            <ArtSvgIcon icon="ri:refresh-line" />
          </ElButton>
          <ElButton
            v-auth="'pay:channel:create'"
            type="primary"
            @click="showDialog('add')"
            v-ripple
          >
            <ArtSvgIcon icon="ri:add-line" class="mr-1" />
            新增渠道
          </ElButton>
        </ElSpace>
      </div>

      <div v-loading="loading" class="channel-body">
        <ElEmpty v-if="!filteredChannels.length" description="暂无支付渠道">
          <ElButton v-auth="'pay:channel:create'" type="primary" @click="showDialog('add')">
            新增第一个渠道
          </ElButton>
        </ElEmpty>

        <div v-else class="channel-grid">
          <div
            v-for="channel in filteredChannels"
            :key="channel.id"
            class="channel-card"
            :class="{ 'is-off': channel.status !== 1 }"
          >
            <div class="channel-card__head">
              <div class="channel-logo" :class="channel.channel === 'wx' ? 'is-wx' : 'is-alipay'">
                <ArtSvgIcon
                  :icon="channel.channel === 'wx' ? 'ri:wechat-pay-line' : 'ri:alipay-line'"
                />
              </div>
              <div class="channel-title">
                <p class="channel-name" :title="channel.name">{{ channel.name }}</p>
                <p class="channel-code">{{ CODE_LABELS[channel.code] || channel.code }}</p>
              </div>
              <ElTag :type="channel.status === 1 ? 'success' : 'info'" size="small" effect="light">
                {{ channel.status === 1 ? '已启用' : '已停用' }}
              </ElTag>
            </div>

            <div class="channel-card__body">
              <div class="channel-meta">
                <span class="meta-label">渠道编码</span>
                <span class="meta-value">{{ channel.code }}</span>
              </div>
              <div class="channel-meta">
                <span class="meta-label">支付方式</span>
                <span class="meta-value">
                  {{ METHOD_LABELS[channel.payMethod] || channel.payMethod }}
                </span>
              </div>
              <div class="channel-meta">
                <span class="meta-label">排序</span>
                <span class="meta-value">{{ channel.sort }}</span>
              </div>
            </div>

            <p class="channel-remark" :title="channel.remark">
              {{ channel.remark || '暂无备注' }}
            </p>

            <div class="channel-card__footer">
              <ElButton
                v-auth="'pay:channel:update'"
                link
                type="primary"
                @click="showDialog('edit', channel)"
              >
                编辑
              </ElButton>
              <ElButton v-auth="'pay:channel:update'" link @click="handleTest(channel)">
                测通
              </ElButton>
              <ElButton
                v-auth="'pay:channel:update'"
                link
                :type="channel.status === 1 ? 'warning' : 'success'"
                @click="handleToggleStatus(channel)"
              >
                {{ channel.status === 1 ? '停用' : '启用' }}
              </ElButton>
              <ElButton
                v-auth="'pay:channel:delete'"
                link
                type="danger"
                @click="handleDelete(channel)"
              >
                删除
              </ElButton>
            </div>
          </div>
        </div>
      </div>
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
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import {
    fetchDeletePayChannel,
    fetchGetPayChannelList,
    fetchSetPayChannelStatus,
    fetchTestPayChannel
  } from '@/api/pay/channel'
  import {
    ElButton,
    ElCard,
    ElEmpty,
    ElMessage,
    ElMessageBox,
    ElRadioButton,
    ElRadioGroup,
    ElSpace,
    ElTag
  } from 'element-plus'
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
  // 本地厂商筛选：渠道总数很少（9 种），一次取全量、前端切换即可
  const activeVendor = ref('')

  const {
    data,
    loading,
    getData,
    searchParams,
    resetSearchParams,
    refreshData,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetPayChannelList,
      apiParams: {
        pageNo: 1,
        pageSize: 100,
        ...searchForm.value
      },
      // 卡片布局不展示表格列，保留空定义只为复用 useTable 的加载与刷新能力
      columnsFactory: () => []
    }
  })

  // useTable 只对 records 结构做类型推导，这里的接口是 { list, total }，显式断言回渠道类型
  const filteredChannels = computed<ChannelItem[]>(() => {
    const list = data.value as unknown as ChannelItem[]

    return activeVendor.value
      ? list.filter((channel) => channel.channel === activeVendor.value)
      : list
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
  const handleToggleStatus = async (row: ChannelItem): Promise<void> => {
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

<style lang="scss" scoped>
  .channel-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 16px;
    flex-wrap: wrap;
  }

  .channel-body {
    min-height: 220px;
  }

  .channel-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
  }

  .channel-card {
    display: flex;
    flex-direction: column;
    padding: 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    background: var(--el-bg-color);
    transition:
      box-shadow 0.2s,
      border-color 0.2s,
      transform 0.2s;

    &:hover {
      border-color: var(--el-color-primary-light-5);
      box-shadow: 0 6px 18px rgb(0 0 0 / 6%);
      transform: translateY(-2px);
    }

    &.is-off {
      background: var(--el-fill-color-lighter);

      .channel-logo {
        filter: grayscale(1);
        opacity: 0.6;
      }
    }
  }

  .channel-card__head {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .channel-logo {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 10px;
    font-size: 22px;
    color: #fff;
    flex-shrink: 0;

    &.is-wx {
      background: linear-gradient(135deg, #2bde6b, #09bb07);
    }

    &.is-alipay {
      background: linear-gradient(135deg, #4a9bff, #1677ff);
    }
  }

  .channel-title {
    flex: 1;
    min-width: 0;
  }

  .channel-name {
    margin: 0;
    overflow: hidden;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .channel-code {
    margin: 2px 0 0;
    overflow: hidden;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .channel-card__body {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 20px;
    margin-top: 14px;
    padding-top: 12px;
    border-top: 1px dashed var(--el-border-color-lighter);
  }

  .channel-meta {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .meta-label {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .meta-value {
    font-size: 13px;
    color: var(--el-text-color-regular);
  }

  .channel-remark {
    display: -webkit-box;
    margin: 12px 0 0;
    overflow: hidden;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }

  .channel-card__footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 4px;
    margin-top: auto;
    padding-top: 12px;
  }
</style>
