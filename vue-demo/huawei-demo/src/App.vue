<template>
  <a-layout style="min-height: 100vh">
    <!-- 侧边栏 -->
    <!-- 侧边栏 - 修改为藏青色背景 -->
    <a-layout-sider :theme="'dark'" width="240" style="background-color: #0f2c59;">
      <div style="color: white;height: 32px; margin: 16px; font-weight: bold;">华为原型系统</div>

      <a-menu
          mode="inline"
          :selectedKeys="[selectedKey]"
          @click="handleMenuClick"
          style="background-color: #0f2c59; color: white;"
          :theme="'dark'"
      >
        <a-menu-item key="home" style="font-size: 14px;">
          <template #icon><i class="fas fa-home"></i></template>
          系统概览
        </a-menu-item>
        <a-menu-item key="request" style="font-size: 14px;">
          <template #icon><i class="fas fa-certificate"></i></template>
          证书申请
        </a-menu-item>
        <a-menu-item key="issue" style="font-size: 14px;">
          <template #icon><i class="fas fa-file-signature"></i></template>
          证书签发
        </a-menu-item>
        <a-menu-item key="renew" style="font-size: 14px;">
          <template #icon><i class="fas fa-sync-alt"></i></template>
          证书更新
        </a-menu-item>
        <a-menu-item key="revoke" style="font-size: 14px;">
          <template #icon><i class="fas fa-ban"></i></template>
          证书撤销
        </a-menu-item>
        <a-menu-item key="query" style="font-size: 14px;">
          <template #icon><i class="fas fa-search"></i></template>
          证书查询
        </a-menu-item>
        <a-menu-item key="verify" style="font-size: 14px;">
          <template #icon><i class="fas fa-check-circle"></i></template>
          证书验证
        </a-menu-item>
        <a-menu-item key="about" style="font-size: 14px;">
          <template #icon><i class="fas fa-info-circle"></i></template>
          关于系统
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <!-- 内容区 -->
    <a-layout-content style="padding: 24px; background-color: #f5f7fa;">
      <router-view />
    </a-layout-content>
  </a-layout>
</template>

<script>
export default {
  data() {
    return {
      selectedKey: this.$route.name, // 当前选中菜单
    };
  },
  watch: {
    // 路由变化时更新选中菜单
    $route(to) {
      this.selectedKey = to.name;
    }
  },
  methods: {
    handleMenuClick({ key }) {
      this.selectedKey = key;
      this.$router.push({ name: key }); // 使用路由名称跳转
    }
  }
};
</script>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css');

/* 菜单项悬停和选中样式 */
.ant-menu-dark .ant-menu-item:hover,
.ant-menu-dark .ant-menu-item-selected {
  background-color: rgba(255, 255, 255, 0.1) !important;
  border-left: 3px solid #4CAF50 !important;
}

.ant-menu-dark .ant-menu-item {
  margin: 4px 0 !important;
  padding-left: 24px !important;
}

/* 菜单图标样式 */
.ant-menu-item .anticon {
  margin-right: 12px !important;
  font-size: 14px !important;
}
</style>
