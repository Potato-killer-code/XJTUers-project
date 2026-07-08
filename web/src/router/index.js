import { createRouter, createWebHistory } from 'vue-router'
import StoreView from '../views/StoreView.vue'
import RetrieveView from '../views/RetrieveView.vue'

const routes = [
  { path: '/', redirect: '/store' },
  { path: '/store', name: 'store', component: StoreView },
  { path: '/retrieve', name: 'retrieve', component: RetrieveView },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
