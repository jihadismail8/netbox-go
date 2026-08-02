import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { hasPermission, type Action } from '@/composables/usePermissions'
import { CORE_RESOURCE_REGISTRY } from './core-resource-registry'

import DefaultLayout from '@/layouts/DefaultLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'

const CoreResourceListView = () => import('@/pages/core/CoreResourceListView.vue')
const CoreResourceDetailView = () => import('@/pages/core/CoreResourceDetailView.vue')
const CoreResourceEditView = () => import('@/pages/core/CoreResourceEditView.vue')
const CoreResourceDeleteView = () => import('@/pages/core/CoreResourceDeleteView.vue')
const DashboardView = () => import('@/pages/special/DashboardView.vue')
const ApiDocsView = () => import('@/pages/special/ApiDocsView.vue')
const LoginPage = () => import('@/pages/auth/LoginPage.vue')
const LogoutAction = () => import('@/pages/auth/LogoutAction.vue')
const NotFoundView = () => import('@/pages/errors/NotFoundView.vue')
const ForbiddenView = () => import('@/pages/errors/ForbiddenView.vue')
const ServerErrorView = () => import('@/pages/errors/ServerErrorView.vue')

const modelRoutes: RouteRecordRaw[] = CORE_RESOURCE_REGISTRY.flatMap((config) => [
  {
    path: config.routePath.slice(1),
    component: CoreResourceListView,
    name: `${config.model}-list`,
    props: { config },
    meta: { permission: { action: 'view', app: config.module, model: config.model } },
  },
  {
    path: `${config.routePath}add/`.slice(1),
    component: CoreResourceEditView,
    name: `${config.model}-add`,
    props: { config },
    meta: { permission: { action: 'add', app: config.module, model: config.model } },
  },
  {
    path: `${config.routePath}:id(\\d+)/`.slice(1),
    component: CoreResourceDetailView,
    name: `${config.model}-detail`,
    props: { config },
    meta: { permission: { action: 'view', app: config.module, model: config.model } },
  },
  {
    path: `${config.routePath}:id(\\d+)/edit/`.slice(1),
    component: CoreResourceEditView,
    name: `${config.model}-edit`,
    props: { config },
    meta: { permission: { action: 'change', app: config.module, model: config.model } },
  },
  {
    path: `${config.routePath}:id(\\d+)/delete/`.slice(1),
    component: CoreResourceDeleteView,
    name: `${config.model}-delete`,
    props: { config },
    meta: { permission: { action: 'delete', app: config.module, model: config.model } },
  },
])

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    component: AuthLayout,
    children: [{ path: '', name: 'login', component: LoginPage }],
  },
  {
    path: '/',
    component: DefaultLayout,
    children: [
      { path: '', name: 'dashboard', component: DashboardView },
      { path: 'api/', name: 'api-docs', component: ApiDocsView },
      { path: 'logout/', name: 'logout', component: LogoutAction },
      { path: '403/', name: 'forbidden', component: ForbiddenView },
      { path: '500/', name: 'server-error', component: ServerErrorView },
      ...modelRoutes,
      { path: ':pathMatch(.*)*', name: 'not-found', component: NotFoundView },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  await authStore.restoreSession()

  if (to.name === 'login' && authStore.isAuthenticated) {
    return { name: 'dashboard' }
  }
  if (to.name !== 'login' && !authStore.isAuthenticated) {
    return { name: 'login', query: { next: to.fullPath } }
  }

  const required = to.meta.permission as { action: Action; app: string; model: string } | undefined
  if (
    required &&
    !hasPermission(
      authStore.user,
      authStore.permissions,
      required.action,
      required.app,
      required.model,
    )
  ) {
    return { name: 'forbidden' }
  }

  return true
})

export default router
