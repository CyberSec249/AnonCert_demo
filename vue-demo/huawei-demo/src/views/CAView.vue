<template>
  <a-card title="CA 列表" :bordered="false">
    <!-- 工具栏：左上角 注册 CA 按钮 -->
    <div class="table-toolbar">
      <a-button type="primary" @click="submitCreate">
        <i class="fa-solid fa-user-plus" style="margin-right:6px;"></i>
        注册 CA
      </a-button>
    </div>

    <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        rowKey="ca_id"
        bordered
        size="middle"
        table-layout="fixed"
        :scroll="{ x: 860 }"
    >
      <template #bodyCell="{ column, record }">
        <!-- 公钥列：省略 + tooltip + 可复制 -->
        <template v-if="column.key === 'caPublicKey'">
          <a-typography-text :ellipsis="{ tooltip: record.caPublicKey }" :copyable="{text: String(record.caPublicKey) || ''}">
            {{ formatKey(record.caPublicKey) }}
          </a-typography-text>
        </template>

        <!-- 状态列 -->
        <template v-else-if="column.key === 'status'">
          <a-tag :color="statusColor(record.caStatus)">{{ statusText(record.caStatus) }}</a-tag>
        </template>

        <!-- 操作列 -->
        <template v-else-if="column.key === 'actions'">
          <a-space size="small" style="flex-wrap:wrap;">
            <a-button size="small" @click="onViewCert(record)">
              <i class="fa-solid fa-eye" style="margin-right:6px;"></i>查看证书
            </a-button>
            <a-button size="small" type="primary" @click="onDownloadCert(record.ca_id)">
              <i class="fa-solid fa-download" style="margin-right:6px;"></i>下载证书
            </a-button>
          </a-space>
        </template>
      </template>
    </a-table>

    <!-- 查看证书弹窗 -->
    <a-modal v-model:open="viewOpen" title="证书详情" width="720px" :footer="null">
      <a-typography-paragraph copyable code style="max-height:360px; overflow:auto; white-space:pre-wrap;">
        {{ selectedCertPem || '（暂无证书数据）' }}
      </a-typography-paragraph>
    </a-modal>

  </a-card>
</template>

<script lang="ts">
import { message } from 'ant-design-vue'

export default {
  data() {
    return {
      loading: false,  // 控制表格加载状态
      columns: [
        { title: 'CA 公钥', dataIndex: 'caPublicKey', key: 'caPublicKey', align: 'center', width: 400, ellipsis: true },
        { title: 'CA 状态', dataIndex: 'status', key: 'status', align: 'center', width: 120 },
        { title: '操作', key: 'actions', align: 'center', width: 220 },
      ],
      dataSource: [],  // 存储从后端加载的 CA 数据
      viewOpen: false,
      selectedCertPem: '',

      // 注册 CA 弹窗相关
      createOpen: false,
      createLoading: false,
      createForm: {
        caPublicKey: '',
        caStatus: 'active',
        ca_id: '',
      },
      statusOptions: [
        { label: '启用', value: 'active' },
        { label: '停用', value: 'suspended' },
        { label: '已撤销', value: 'revoked' },
      ],
    }
  },
  methods: {
    formatKey(key) {
      if (!key) return ''
      const s = String(key)
      return s.length > 24 ? `${s.slice(0, 10)}...${s.slice(-8)}` : s
    },
    statusColor(status) {
      switch (status) {
        case 'active': return 'green'
        case 'suspended': return 'orange'
        case 'revoked': return 'red'
        default: return 'default'
      }
    },
    statusText(status) {
      switch (status) {
        case 'active': return '启用'
        case 'suspended': return '停用'
        case 'revoked': return '已撤销'
        default: return status || '未知'
      }
    },
    onViewCert(record) {
      this.selectedCertPem = record.certPem || ''
      if (!this.selectedCertPem) {
        message.info('该 CA 暂无证书 PEM')
      }
      this.viewOpen = true
    },

    async onDownloadCert(caID: number | string) {
      const url = `/api/ca/cert/download?id=${encodeURIComponent(String(caID))}`
      window.open(url, '_blank')
    },

    // ——— 注册 CA ———
    async submitCreate() {
      this.createLoading = true
      try {
        const res = await fetch('/api/ca/register', { method:'POST' })
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        message.success('注册成功')

        await this.fetchList()
      } catch (e) {
        console.error(e)
        message.error('注册失败，请稍后重试')
      } finally {
        this.createLoading = false
      }
    },

    // —— 列表加载
    async fetchList() {
      this.loading = true
      try {
        const res = await fetch('/api/ca/list')
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const data = await res.json()

        this.dataSource = Array.isArray(data.items) ? data.items : []
      } catch (e) {
        console.error(e)
        message.error('加载证书列表失败')
      } finally {
        this.loading = false
      }
    }
  },
  mounted() {
    this.fetchList() // 页面加载时拉取 CA 列表
  }
}
</script>

<style>
.table-toolbar {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 公钥列用等宽字体（可选） */
.ant-table td:nth-child(1) .ant-typography {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

/* 表格内按钮更紧凑 */
.ant-table .ant-btn {
  padding: 0 10px;
}

</style>
