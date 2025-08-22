<template>
  <a-layout style="min-height: 100vh">
    <!-- 侧边栏 -->
    <a-layout-sider :theme="'dark'" width="240" style="background-color: #0f2c59;">
      <div class="logo">华为原型系统</div>

      <a-menu
          mode="inline"
          :selectedKeys="[selectedKey]"
          :openKeys="openKeys"
          @openChange="onOpenChange"
          @click="handleMenuClick"
          :theme="'dark'"
          style="background-color: #0f2c59; color: white;"
      >
        <!-- 模块 1：系统概览 -->
        <a-menu-item key="home" style="font-size: 14px;">
          <template #icon><i class="fa-solid fa-home"></i></template>
          系统概览
        </a-menu-item>

        <!-- 模块 2：CA 功能（折叠） -->
        <a-sub-menu key="caGroup">
          <template #title>
            <span class="group-title">
              <i class="fa-solid fa-building-columns group-icon"></i>
              CA 功能
            </span>
          </template>

          <a-menu-item key="issue" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-file-signature"></i></template>
            证书签发
          </a-menu-item>

          <a-menu-item key="renew" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-rotate"></i></template>
            证书更新
          </a-menu-item>

          <a-menu-item key="revoke" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-ban"></i></template>
            证书撤销
          </a-menu-item>
        </a-sub-menu>

        <!-- 模块 3：用户功能（折叠） -->
        <a-sub-menu key="userGroup">
          <template #title>
            <span class="group-title">
              <i class="fa-solid fa-user-shield group-icon"></i>
              用户功能
            </span>
          </template>

          <a-menu-item key="ca" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-building-columns"></i></template>
            CA 列表
          </a-menu-item>

          <a-menu-item key="request" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-certificate"></i></template>
            证书申请
          </a-menu-item>

          <a-menu-item key="query" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-magnifying-glass"></i></template>
            证书查询
          </a-menu-item>

          <a-menu-item key="verify" style="font-size: 14px;">
            <template #icon><i class="fa-solid fa-circle-check"></i></template>
            证书验证
          </a-menu-item>
        </a-sub-menu>

        <!-- 模块 4：关于系统 -->
        <a-menu-item key="about" style="font-size: 14px;">
          <template #icon><i class="fa-solid fa-circle-info"></i></template>
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
      selectedKey: this.$route.name,                    // 当前选中
      openKeys: this.computeOpenKeys('caGroup', 'userGroup'), // 当前展开
    };
  },
  watch: {
    $route(to) {
      this.selectedKey = to.name;
      this.openKeys = this.computeOpenKeys(to.name);
    }
  },
  methods: {
    handleMenuClick({ key }) {
      this.selectedKey = key;
      this.$router.push({ name: key });

      // 点击子项时，确保对应分组展开
      const needOpenCA = ['issue', 'renew', 'revoke'].includes(key);
      const needOpenUser = ['ca', 'request', 'query', 'verify'].includes(key);

      const next = new Set(this.openKeys);
      if (needOpenCA) next.add('caGroup');
      if (needOpenUser) next.add('userGroup');
      this.openKeys = Array.from(next);
    },

    computeOpenKeys(routeName) {
      const res = [];
      if (['issue', 'renew', 'revoke'].includes(routeName)) res.push('caGroup');
      if (['ca', 'request', 'query', 'verify'].includes(routeName)) res.push('userGroup');
      return ['caGroup', 'userGroup'];
    },
  }
};
</script>

<style>
@import url('https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css');

.logo {
  color: #fff;
  height: 32px;
  margin: 16px;
  font-weight: bold;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 分组/子菜单标题样式 */
.group-title {
  color: rgba(255,255,255,0.85);
  font-weight: 600;
  font-size: 13px;
  letter-spacing: 0.5px;
  display: inline-flex;
  align-items: center;
}
.group-icon {
  margin-right: 8px;
  font-size: 14px;
}

/* 悬停与选中高亮 */
.ant-menu-dark .ant-menu-item:hover,
.ant-menu-dark .ant-menu-item-selected {
  background-color: rgba(255, 255, 255, 0.1) !important;
  border-left: 3px solid #4CAF50 !important;
}

/* 行距与缩进 */
.ant-menu-dark .ant-menu-item,
.ant-menu-dark .ant-menu-submenu-title {
  margin: 4px 0 !important;
  padding-left: 24px !important;
}

/* 图标间距 */
.ant-menu-item .anticon,
.ant-menu-submenu-title .anticon {
  margin-right: 12px !important;
  font-size: 14px !important;
}

/* 子菜单箭头颜色 */
.ant-menu-dark .ant-menu-submenu-title .ant-menu-submenu-expand-icon,
.ant-menu-dark .ant-menu-submenu-title .ant-menu-submenu-arrow {
  color: rgba(255,255,255,0.65);
}
</style>