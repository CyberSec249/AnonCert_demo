<template>
  <div class="home-container">
    <!-- 欢迎标题 -->
    <div class="welcome-section">
      <h1 class="welcome-title">欢迎使用DPKI原型系统！</h1>
      <p class="welcome-text">
        这是华为基于区块链技术的证书管理系统WEB界面。欢迎您的使用！左侧导航栏提供了不同操作功能的快捷访问，您可以选择切换到证书签发、更新、撤销、查询、验证功能，所有流程均可以在关于系统页面获得技术支持！
      </p>
    </div>

    <!-- 客户端配置与工作情况 -->
    <div class="config-stats-container">
      <!-- 客户端配置 (占3/4宽度) -->
      <div class="config-section">
        <div class="section-header">
          <div class="icon-container">
            <i class="fas fa-cog icon"></i>
          </div>
          <h2 class="section-title">DPKI - 区块链网络配置</h2>
        </div>

        <div class="config-grid">
          <div class="config-card">
            <div class="config-label">区块链节点数量</div>
            <div class="config-value">{{nodeCount}}</div>
          </div>

          <div class="config-card">
            <div class="config-label">区块数量</div>
            <div class="config-value">{{blockCount}}</div>
          </div>

          <div class="config-card">
            <div class="config-label">交易数量</div>
            <div class="config-value">{{txCount}}</div>
          </div>

          <div class="config-card">
            <div class="config-label">智能合约数量</div>
            <div class="config-value">{{contractCount}}</div>
          </div>
        </div>

        <div class="carousel-section">
          <div
              class="carousel"
              @mouseenter="pauseCarousel"
              @mouseleave="resumeCarousel"
              ref="carouselRef"
          >
            <div
                class="carousel-track"
                :style="{ transform: `translateX(-${currentSlide * 100}%)` }"
            >
              <!-- 替换为你的实际图片路径或网络图片 -->
              <div class="carousel-slide" v-for="(img, idx) in carouselImages" :key="idx">
                <img :src="img.src" :alt="img.alt" />
                <div class="carousel-caption" v-if="img.caption">
                  {{ img.caption }}
                </div>
              </div>
            </div>

            <!-- 左右按钮 -->
            <button class="carousel-btn prev" @click="prevSlide" aria-label="上一张">
              <i class="fas fa-chevron-left"></i>
            </button>
            <button class="carousel-btn next" @click="nextSlide" aria-label="下一张">
              <i class="fas fa-chevron-right"></i>
            </button>

            <!-- 指示点 -->
            <div class="carousel-dots">
              <button
                  v-for="(img, idx) in carouselImages"
                  :key="'dot-' + idx"
                  class="dot"
                  :class="{ active: idx === currentSlide }"
                  @click="goToSlide(idx)"
                  :aria-label="`跳转到第 ${idx + 1} 张`"
              ></button>
            </div>
          </div>
        </div>
      </div>

      <!-- 工作情况 (占1/4宽度) -->
      <div class="stats-section">
        <div class="section-header">
          <div class="icon-container">
            <i class="fas fa-chart-line icon"></i>
          </div>
          <h2 class="section-title">DPKI - 证书数据</h2>
        </div>

        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-label">待签发证书</div>
            <div class="stat-value">1</div>
          </div>

          <div class="stat-card">
            <div class="stat-label">已签发证书</div>
            <div class="stat-value">4</div>
          </div>

          <div class="stat-card">
            <div class="stat-label">已吊销证书</div>
            <div class="stat-value">4</div>
          </div>

          <div class="stat-card">
            <div class="stat-label">已更新证书</div>
            <div class="stat-value">7</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统功能卡片 - 2+3布局 -->
    <div class="features-section">
      <div class="section-header">
        <div class="icon-container">
          <i class="fas fa-cube icon"></i>
        </div>
        <h2 class="section-title">系统核心功能</h2>
      </div>

      <div class="features-grid">
        <div class="feature-card">
          <div class="feature-icon-container">
            <i class="fas fa-key feature-icon"></i>
          </div>
          <h3 class="feature-title">证书签发</h3>
          <p class="feature-desc">通过区块链技术实现去中心化的证书签发流程</p>
        </div>

        <div class="feature-card">
          <div class="feature-icon-container">
            <i class="fas fa-sync feature-icon"></i>
          </div>
          <h3 class="feature-title">证书更新</h3>
          <p class="feature-desc">自动化证书更新流程，减少人工干预</p>
        </div>

        <div class="feature-card">
          <div class="feature-icon-container">
            <i class="fas fa-ban feature-icon"></i>
          </div>
          <h3 class="feature-title">证书撤销</h3>
          <p class="feature-desc">实时撤销无效证书，确保系统安全</p>
        </div>

        <div class="feature-card">
          <div class="feature-icon-container">
            <i class="fas fa-search feature-icon"></i>
          </div>
          <h3 class="feature-title">证书查询</h3>
          <p class="feature-desc">提供快速高效的证书查询功能</p>
        </div>

        <div class="feature-card">
          <div class="feature-icon-container">
            <i class="fas fa-check-circle feature-icon"></i>
          </div>
          <h3 class="feature-title">证书验证</h3>
          <p class="feature-desc">实时验证证书有效性，确保业务安全</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import img1 from '@/assets/1.jpeg';
import img2 from '@/assets/2.jpeg';
import img3 from '@/assets/3.jpg';

export default defineComponent({
  name: 'HomePage',
  data() {
    return {
      // 区块链统计（初始化）
      nodeCount: 0,
      blockCount: 0,
      txCount: 0,
      contractCount: 0,

      // 轮播相关
      currentSlide: 0,
      carouselTimer: null,
      carouselIntervalMs: 4000,
      carouselImages: [
        { src: img3, alt: '公钥基础设施', caption: '公钥基础设施' },
        { src: img2, alt: '链上审计与查询', caption: '链上审计与查询' },
        { src: img1, alt: '证书生命周期', caption: '证书生命周期' }
      ],
    };
  },
  methods: {
    nextSlide() {
      this.currentSlide = (this.currentSlide + 1) % this.carouselImages.length;
    },
    prevSlide() {
      this.currentSlide = (this.currentSlide - 1 + this.carouselImages.length) % this.carouselImages.length;
    },
    goToSlide(idx) {
      this.currentSlide = idx;
    },
    startCarousel() {
      this.stopCarousel();
      this.carouselTimer = setInterval(this.nextSlide, this.carouselIntervalMs);
    },
    stopCarousel() {
      if (this.carouselTimer) {
        clearInterval(this.carouselTimer);
        this.carouselTimer = null;
      }
    },
    pauseCarousel() { this.stopCarousel(); },
    resumeCarousel() { this.startCarousel(); },

    // 移动端滑动
    attachTouchHandlers() {
      const el = this.$refs.carouselRef; // 统一用 $refs
      if (!el) return;
      let startX = 0, deltaX = 0;

      const onTouchStart = (e) => {
        startX = e.touches[0].clientX;
        deltaX = 0;
        this.pauseCarousel();
      };
      const onTouchMove = (e) => {
        deltaX = e.touches[0].clientX - startX;
      };
      const onTouchEnd = () => {
        if (Math.abs(deltaX) > 50) {
          deltaX < 0 ? this.nextSlide() : this.prevSlide();
        }
        this.resumeCarousel();
      };

      el.addEventListener('touchstart', onTouchStart, { passive: true });
      el.addEventListener('touchmove', onTouchMove, { passive: true });
      el.addEventListener('touchend', onTouchEnd);
      this._touchHandlers = { onTouchStart, onTouchMove, onTouchEnd, el };
    },
    detachTouchHandlers() {
      const h = this._touchHandlers;
      if (!h) return;
      h.el.removeEventListener('touchstart', h.onTouchStart);
      h.el.removeEventListener('touchmove', h.onTouchMove);
      h.el.removeEventListener('touchend', h.onTouchEnd);
      this._touchHandlers = null;
    },
  },
  async mounted() {         // 注意这里就用一个空格
    this.startCarousel();
    this.attachTouchHandlers();

    try {
      const res = await fetch('/api/metrics'); // 依赖 vite 代理
      const data = await res.json();
      this.nodeCount     = data.nodeCount ?? 0;
      this.blockCount    = data.blockCount ?? 0;
      this.txCount       = data.txCount ?? 0;
      this.contractCount = data.contractCount ?? 0;
    } catch (e) {
      console.error('加载指标失败：', e);
    }
  },
  beforeUnmount() {
    this.stopCarousel();
    this.detachTouchHandlers();
  },
});
</script>
<style scoped>
.home-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 16px;
}

/* 欢迎部分样式 */
.welcome-section {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.welcome-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--huawei-blue);
  margin-bottom: 16px;
  text-align: left;
}

.welcome-text {
  font-size: 16px;
  line-height: 1.6;
  color: #333;
  text-align: left;
  margin-bottom: 0;
}

/* 配置和统计容器 */
.config-stats-container {
  display: grid;
  grid-template-columns: 2fr 1.2fr; /* 右边加宽 */
  gap: 24px;
  margin-bottom: 24px;
}

/* 配置部分样式 */
.config-section, .stats-section, .features-section {
  background-color: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.section-header {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #eee;
}

.icon-container {
  background-color: rgba(15, 44, 89, 0.1);
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
}

.icon {
  font-size: 20px;
  color: var(--huawei-blue);
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--huawei-blue);
  margin: 0;
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.config-item {
  margin-bottom: 16px;
}

.config-item.full-width {
  grid-column: span 2;
}

.config-label {
  font-size: 16px;
  color: #333;
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
}

.config-value {
  font-size: 18px;
  color: #666;
  margin: 0;
  padding-left: 8px;
}

.contract-address {
  word-break: break-all;
  font-family: monospace;
  font-size: 18px;
}

/* 统计部分样式 */
.stats-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

.stat-card {
  background: linear-gradient(135deg, #4a6585 0%, #2c3e50 100%);
  border-radius: 8px;
  padding: 16px;
  color: white;
}

.stat-label {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
}

/* 功能卡片部分 */
.features-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
}

.feature-card {
  text-align: center;
  padding: 20px;
  border: 1px solid #eee;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.feature-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.1);
  border-color: var(--huawei-blue);
}

.feature-icon-container {
  background-color: rgba(15, 44, 89, 0.1);
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.feature-icon {
  font-size: 24px;
  color: var(--huawei-blue);
}

.feature-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--huawei-blue);
  margin-bottom: 8px;
}

.feature-desc {
  font-size: 14px;
  color: #666;
  margin: 0;
}

.config-card {
  background: linear-gradient(135deg, #e9f0fa 0%, #f8fbff 100%);
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.05);
  text-align: center;
}

.config-card .config-label {
  font-size: 14px;
  font-weight: 500;
  color: #555;
  margin-bottom: 6px;
}

.config-card .config-value {
  font-size: 22px;
  font-weight: 700;
  color: var(--huawei-blue);
}

/* 轮播模块 */
.carousel-section {
  margin-top: 35px;
}

.carousel {
  position: relative;
  overflow: hidden;
  border-radius: 10px;
  background: #f6f8fb;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}

.carousel-track {
  display: flex;
  transition: transform 0.6s ease;
  will-change: transform;
}

.carousel-slide {
  min-width: 100%;
  position: relative;
  height: 220px; /* 你也可以用 28vh 等相对高度 */
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eef3fa;
}

.carousel-slide img {
  width: 100%;
  height: 100%;
  object-fit: cover; /* 保持铺满且不变形 */
}

.carousel-caption {
  position: absolute;
  left: 12px;
  bottom: 15px;
  padding: 6px 10px;
  border-radius: 6px;
  background: rgba(15,44,89,0.75);
  color: #fff;
  font-size: 12px;
}

/* 左右切换按钮 */
.carousel-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  border: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255,255,255,0.9);
  color: var(--huawei-blue);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 1px 6px rgba(0,0,0,0.12);
}

.carousel-btn:hover {
  background: #fff;
}

.carousel-btn.prev { left: 10px; }
.carousel-btn.next { right: 10px; }

/* 指示点 */
.carousel-dots {
  position: absolute;
  left: 0; right: 0; bottom: 10px;
  display: flex;
  gap: 8px;
  justify-content: center;
}

.carousel-dots .dot {
  width: 8px;
  height: 8px;
  border: 0;
  border-radius: 50%;
  background: rgba(15,44,89,0.25);
  cursor: pointer;
}

.carousel-dots .dot.active {
  background: var(--huawei-blue);
}

/* 响应式高度微调 */
@media (max-width: 992px) {
  .carousel-slide { height: 200px; }
}
@media (max-width: 576px) {
  .carousel-slide { height: 180px; }
}


/* 响应式调整 */
@media (max-width: 1200px) {
  .features-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 992px) {
  .config-stats-container {
    grid-template-columns: 1fr;
  }

  .features-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .config-grid {
    grid-template-columns: 1fr;
  }

  .config-item.full-width {
    grid-column: span 1;
  }
}

@media (max-width: 576px) {
  .features-grid {
    grid-template-columns: 1fr;
  }

  .welcome-section,
  .config-section,
  .stats-section,
  .features-section {
    padding: 16px;
  }

  .welcome-title {
    font-size: 20px;
  }

  .section-title {
    font-size: 16px;
  }
}

/* 颜色变量 */
:root {
  --huawei-blue: #0f2c59;
  --huawei-light-blue: #4a6585;
}
</style>