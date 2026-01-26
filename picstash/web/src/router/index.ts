import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/pages/index.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/login.vue'),
    },
    {
      path: '/upload',
      name: 'upload',
      component: () => import('@/pages/upload.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/tags',
      name: 'tags',
      component: () => import('@/pages/tags.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/sync',
      name: 'sync',
      component: () => import('@/pages/sync.vue'),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach((_to, _from, next) => {
  const authStore = useAuthStore()
  authStore.initFromStorage()

  if (_to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (_to.path === '/login' && authStore.isAuthenticated) {
    next('/')
  } else {
    next()
  }
})

export default router
