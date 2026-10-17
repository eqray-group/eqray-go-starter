<template>
  <ElDialog
    v-model="dialogVisible"
    :title="`角色授权 - ${role?.name || ''}`"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElAlert
      v-if="isAdmin"
      type="info"
      :closable="false"
      show-icon
      title="超级管理员拥有全部权限，无需单独授权"
    />

    <div v-else v-loading="loading" class="permission-tree">
      <ElTree
        ref="treeRef"
        :data="treeData"
        show-checkbox
        node-key="id"
        :props="{ label: 'label', children: 'children' }"
        :default-checked-keys="checkedKeys"
        :default-expand-all="false"
      />
    </div>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton
        v-if="!isAdmin"
        type="primary"
        :loading="submitting"
        :disabled="loading"
        @click="handleSubmit"
      >
        保存
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import type { ElTree } from 'element-plus'
  import { fetchAssignRoleMenus, fetchGetRoleMenus } from '@/api/system/role'
  import { fetchGetMenuOptions } from '@/api/system/menu'

  type RoleItem = Api.SystemManage.RoleItem
  type MenuItem = Api.SystemManage.MenuItem

  interface Props {
    visible: boolean
    role?: RoleItem | null
  }
  interface Emits {
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }

  interface TreeOption {
    id: string
    label: string
    children: TreeOption[]
  }

  const props = withDefaults(defineProps<Props>(), { role: null })
  const emit = defineEmits<Emits>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value)
  })

  const isAdmin = computed(() => props.role?.code === 'SUPER_ADMIN')
  const loading = ref(false)
  const submitting = ref(false)
  const treeRef = ref<InstanceType<typeof ElTree>>()
  const treeData = ref<TreeOption[]>([])
  const checkedKeys = ref<string[]>([])
  const leafIDs = ref<Set<string>>(new Set())

  /** 后端菜单树 → ElTree 结构，同时记录叶子节点用于回显 */
  const buildTree = (nodes: MenuItem[]): TreeOption[] =>
    nodes.map((node) => {
      const children = buildTree(node.children || [])
      if (children.length === 0) {
        leafIDs.value.add(node.id)
      }
      return { id: node.id, label: node.name, children }
    })

  const loadData = async (): Promise<void> => {
    if (!props.role || isAdmin.value) return

    loading.value = true
    try {
      const [menus, granted] = await Promise.all([
        fetchGetMenuOptions(),
        fetchGetRoleMenus(props.role.id)
      ])

      leafIDs.value = new Set()
      treeData.value = buildTree(menus)
      // 只回显叶子节点：父节点由 ElTree 根据子节点自动勾选 / 半选，避免整棵子树被误勾
      checkedKeys.value = granted.filter((id) => leafIDs.value.has(id))
    } finally {
      loading.value = false
    }
  }

  watch(
    () => props.visible,
    (visible) => {
      if (visible) loadData()
    },
    { immediate: true }
  )

  /** 保存：勾选 + 半选（父级目录）一起提交，保证父级节点也被授权 */
  const handleSubmit = async (): Promise<void> => {
    if (!props.role || !treeRef.value) return

    const checked = treeRef.value.getCheckedKeys(false) as string[]
    const halfChecked = treeRef.value.getHalfCheckedKeys() as string[]
    const menuIDs = Array.from(new Set([...checked, ...halfChecked]))

    submitting.value = true
    try {
      await fetchAssignRoleMenus(props.role.id, menuIDs)
      ElMessage.success('授权成功')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .permission-tree {
    max-height: 420px;
    overflow: auto;
    padding: 4px;
  }
</style>
