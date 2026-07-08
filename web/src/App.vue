<template>
  <div id="app-root">
    <!-- 顶部导航 -->
    <nav class="navbar">
      <div class="nav-inner">
        <span class="logo">智能外卖柜</span>
        <div class="nav-links">
          <router-link to="/store" class="nav-link" active-class="active">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="nav-icon">
              <path d="M12 5v14M5 12h14"/>
            </svg>
            存入
          </router-link>
          <router-link to="/retrieve" class="nav-link" active-class="active">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="nav-icon">
              <path d="M5 12h14M12 5l-7 7 7 7"/>
            </svg>
            取出
          </router-link>
        </div>
      </div>
    </nav>

    <!-- 页面内容 -->
    <main>
      <router-view />
    </main>
  </div>
</template>

<style>
/* 全局样式 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #f1f5f9;
  color: #1e293b;
  -webkit-font-smoothing: antialiased;
}

/* 导航栏 */
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 60px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(226, 232, 240, 0.8);
  z-index: 100;
}

.nav-inner {
  max-width: 480px;
  margin: 0 auto;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
}

.logo {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
}

.nav-links {
  display: flex;
  gap: 4px;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  text-decoration: none;
  transition: all 0.2s;
}

.nav-link:hover {
  background: #f1f5f9;
  color: #334155;
}

.nav-link.active {
  background: #e0f2fe;
  color: #0284c7;
}

.nav-icon {
  width: 18px;
  height: 18px;
}

/* 主内容区 */
main {
  padding-top: 60px;
}
</style>
