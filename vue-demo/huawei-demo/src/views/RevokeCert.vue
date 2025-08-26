<script setup lang="ts">
import { ref, onMounted, h, resolveComponent, reactive, createVNode} from 'vue'
import { message, Modal, Input} from 'ant-design-vue'


type CertItem = {
  id: number
  publicKey: string
  subjectInfo: string
  ca_id: number
  caPublicKey: string
  status: 'issued' | 'revoked'
  description: string
}

const loading = ref(false)
const dataSource = ref<CertItem[]>([])

const columns = [
  {
    title: '主体公钥',
    dataIndex: 'publicKey',
    key: 'publicKey',
    align: 'center',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => {
      return {
        children: h(resolveComponent('a-typography-text'),
            {
              ellipsis: { tooltip: text },
              copyable: { text }
            },
            { default: () => formatKey(text) }
        )
      }
    },
  },
  {
    title: 'CA 公钥',
    dataIndex: 'caPublicKey',
    key: 'caPublicKey',
    align: 'center',
    ellipsis: true,
    customRender: ({ text }: { text: string }) => {
      return {
        children: h(resolveComponent('a-typography-text'),
            {
              ellipsis: { tooltip: text },
              copyable: { text }
            },
            { default: () => formatKey(text) }
        )
      }
    },
  },
  {
    title: '当前状态',
    dataIndex: 'status',
    key: 'status',
    align: 'center',
    customRender: ({ text }: { text: CertItem['status'] }) => {
      const map: Record<CertItem['status'], { color: string, label: string }> = {
        issued: { color: 'green', label: '已签发' },
        revoked: { color: 'red', label: '已撤销' }
      }
      const m = map[text] || { color: 'gray', label: text }
      return { children: h(resolveComponent('a-tag'), { color: m.color }, { default: () => m.label }) }
    }
  },
  { title: '证书说明', dataIndex: 'description', key: 'description', align: 'center', ellipsis: true },
  {
    title: '操作',
    key: 'action',
    align: 'center',
    width: 200,
    customRender: ({ record }: { record: CertItem }) => {
      return h('div', { style: 'display:flex; gap:8px; justify-content:center;' }, [
        h(resolveComponent('a-button'),
            {
              size: 'small',
              type: 'default',
              onClick: () => handleViewDetail(record)
            },
            { default: () => '查看详情' }
        ),
        h(resolveComponent('a-button'),
            {
              size: 'small',
              type: 'primary',
              danger: true,
              disabled: record.status === 'revoked',
              onClick: () => handleRevoke(record)
            },
            { default: () => '撤销证书' }
        )
      ])
    }
  }
]

// —— 加载列表
async function fetchList() {
  loading.value = true
  try {
    const res = await fetch('/api/revoke/list')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    dataSource.value = Array.isArray(data.items) ? data.items : []
  } catch (e) {
    console.error(e)
    message.error('加载证书列表失败')
  } finally {
    loading.value = false
  }
}

// —— 查看详情
const detailModalOpen = ref(false)
const detailLoading = ref(false)
const detailModel = reactive<Partial<CertItem>>({})

async function handleViewDetail(record: CertItem) {
  detailLoading.value = true
  detailModalOpen.value = true
  try {
    const res = await fetch(`/api/request/detail?id=${record.id}`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    Object.assign(detailModel, data)
  } catch (e) {
    message.error('加载详情失败')
    detailModalOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

// —— 撤销证书
async function handleRevoke(record: CertItem) {
  const inputValue = ref('') // 绑定唯一解输入框

  Modal.confirm({
    title: '确认撤销',
    content: () =>
        createVNode('div', {}, [
          createVNode('p', null, `确定要撤销证书 (ID=${record.id}) 吗？`),
          createVNode(Input, {
            style: 'margin-top:8px;',
            placeholder: '请输入唯一解 X',
            type: 'password',            // 👈 密文显示更安全
            onInput: (e: any) => {
              inputValue.value = e.target.value
            }
          })
        ]),
    okText: '确认',
    cancelText: '取消',
    async onOk() {
      if (!inputValue.value) {
        message.warning('请输入唯一解 X')
        return Promise.reject() // 阻止关闭弹窗
      }
      try {
        const res = await fetch('/api/revoke/cert', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: record.id,
            x: inputValue.value.trim()
          })
        })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()
        message.success(data?.msg || '证书已撤销')
        await fetchList()
      } catch (e) {
        console.error(e)
        message.error('撤销失败，请稍后重试')
      }
    }
  })
}

onMounted(fetchList)

// —— 公钥缩略显示
function formatKey(key: string) {
  if (!key) return ''
  return key.length > 30 ? `${key.slice(0, 12)}...${key.slice(-10)}` : key
}
</script>

<template>
  <a-card title="证书撤销列表">
    <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        row-key="id"
        bordered
        :pagination="{ pageSize: 10 }"
    />
  </a-card>

  <!-- 详情弹窗 -->
  <a-modal
      v-model:open="detailModalOpen"
      title="申请信息"
      :footer="null"
      :confirmLoading="detailLoading"
      destroyOnClose
  >
    <a-form layout="vertical">
      <a-form-item label="ID" hidden="hidden">
        <a-input v-model:value="detailModel.id" readonly />
      </a-form-item>

      <a-form-item label="公钥">
        <a-textarea v-model:value="detailModel.publicKey" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

      <a-form-item label="主体信息">
        <a-textarea v-model:value="detailModel.subjectInfo" :auto-size="{ minRows: 2, maxRows: 6 }" readonly />
      </a-form-item>

      <a-form-item label="CA">
        <a-textarea v-model:value="detailModel.caPublicKey" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

      <a-form-item label="证书说明">
        <a-textarea v-model:value="detailModel.description" :auto-size="{ minRows: 2, maxRows: 4 }" readonly />
      </a-form-item>

    </a-form>
  </a-modal>
</template>

<style scoped>
</style>
