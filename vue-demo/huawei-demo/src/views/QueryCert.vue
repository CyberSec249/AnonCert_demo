<template>
  <a-card title="证书查询">
    <a-table
        :columns="columns"
        :data-source="dataSource"
        :loading="loading"
        row-key="cert_id"
        bordered
        size="middle"
        :pagination="{ pageSize: 10 }"
    />
  </a-card>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, h, resolveComponent } from 'vue';
import { message } from 'ant-design-vue';

// 表格项类型
type CertItem = {
  cert_id: number;
  serialHex: string;
  ca_id: number;
  caPublicKey: string;
  cert_path: string; // 签发时间
  timestamp: string;
};

const loading = ref(false);
const dataSource = ref<CertItem[]>([]);

// 列定义
const columns = [
  {
    title: '序列号',
    dataIndex: 'serialHex',
    key: 'serialHex',
    align: 'center',
    width: 220,
  },
  {
    title: '签发CA',
    dataIndex: 'caPublicKey',
    key: 'caPublicKey',
    align: 'center',
    width: 300,
    customRender: ({ text }: { text: string }) => {
      return {
        children: h(
            resolveComponent('a-typography-text'),
            {
              ellipsis: { tooltip: text },
              copyable: { text },
            },
            { default: () => formatKey(text) }
        )
      }
    },
  },
  {
    title: '签发时间',
    dataIndex: 'timestamp',
    key: 'timestamp',
    align: 'center',
    width: 180,
  },
  {
    title: '操作',
    key: 'actions',
    align: 'center',
    width: 180,
    customRender: ({ record }: { record: CertItem }) => {
      return h('div', { style: 'display:flex; gap:8px; justify-content:center; width:100%;' }, [
        h(resolveComponent('a-button'),
            {
              size: 'small',
              type: 'primary',
              onClick: () => onDownloadCert(record.cert_id),
            },
            { default: () => '下载证书' }
        )
      ]);
    }
  }
];

// 数据加载
async function fetchList() {
  loading.value = true;
  try {
    const res = await fetch('/api/cert/list');  // 替换为你的实际接口
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    dataSource.value = Array.isArray(data.items) ? data.items : [];
  } catch (e) {
    console.error(e);
    message.error('加载证书列表失败');
  } finally {
    loading.value = false;
  }
}

async function onDownloadCert(cert_id: number | string) {
  const url = `/api/cert/download?cert_id=${encodeURIComponent(String(cert_id))}`
  window.open(url, '_blank')
}


// 格式化公钥
function formatKey(key: string) {
  if (!key) return '';
  const s = String(key);
  return s.length > 30 ? `${s.slice(0, 12)}...${s.slice(-10)}` : s;
}

// 页面加载时查询数据
onMounted(fetchList);
</script>

<style scoped>
/* 你可以根据实际需求自定义样式 */
</style>
