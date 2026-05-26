import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { getToken } from '@/utils/storage'

const routes: RouteRecordRaw[] = [
  {
    path: '/install',
    name: 'Install',
    component: () => import('@/views/install/InstallView.vue'),
    meta: { requiresAuth: false, layout: 'install' },
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { requiresAuth: false, layout: 'auth' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/auth/RegisterView.vue'),
    meta: { requiresAuth: false, layout: 'auth' },
  },
  {
    path: '/',
    redirect: '/spaces',
  },
  {
    path: '/spaces',
    name: 'Spaces',
    component: () => import('@/views/space/SpaceListView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/spaces/:id',
    name: 'SpaceDetail',
    component: () => import('@/views/space/SpaceDetailView.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'pages/:pageId',
        name: 'PageView',
        component: () => import('@/views/page/PageView.vue'),
      },
      {
        path: 'pages/:pageId/edit',
        name: 'PageEditor',
        component: () => import('@/views/page/PageEditorView.vue'),
      },
      {
        path: 'pages/:pageId/versions',
        name: 'VersionHistory',
        component: () => import('@/views/page/VersionHistoryView.vue'),
      },
    ],
  },
  {
    path: '/spaces/:id/settings',
    name: 'SpaceSettings',
    component: () => import('@/views/space/SpaceSettingsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/profile/ProfileView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/AdminWrapper.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'admin' },
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/DashboardView.vue') },
      { path: 'users', name: 'AdminUsers', component: () => import('@/views/admin/UserManageView.vue') },
      { path: 'spaces', name: 'AdminSpaces', component: () => import('@/views/admin/SpaceManageView.vue') },
      { path: 'pages', name: 'AdminPages', component: () => import('@/views/admin/PageManageView.vue') },
      { path: 'settings', name: 'AdminSettings', component: () => import('@/views/admin/SettingsView.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let installChecked = false

router.beforeEach(async (to, _from, next) => {
  if (!installChecked) {
    installChecked = true
    try {
      const { getInstallStatus } = await import('@/api/install')
      const status = await getInstallStatus()
      if (!status.installed && to.path !== '/install') {
        next({ name: 'Install' })
        return
      }
      if (status.installed && to.path === '/install') {
        next({ name: 'Login' })
        return
      }
    } catch {
      // Network error — assume installed to avoid blocking
    }
  }

  const token = getToken()
  if (to.meta.requiresAuth && !token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if ((to.path === '/login' || to.path === '/register') && token) {
    next({ name: 'Spaces' })
  } else {
    next()
  }
})

export default router
