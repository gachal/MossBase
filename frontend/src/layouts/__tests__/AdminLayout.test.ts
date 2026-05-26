import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import AdminLayout from '../AdminLayout.vue'

async function mountWithRouter() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/admin/dashboard', component: { template: '<div />' } },
      { path: '/spaces', name: 'Spaces', component: { template: '<div />' } },
    ],
  })
  await router.push('/admin/dashboard')
  await router.isReady()

  return mount(AdminLayout, {
    global: {
      plugins: [router],
      stubs: {
        'el-container': { template: '<div><slot /></div>' },
        'el-aside': { template: '<div><slot /></div>' },
        'el-main': { template: '<div><slot /></div>' },
        'el-menu': { template: '<div><slot /></div>' },
        'el-menu-item': { template: '<div><slot /></div>', props: ['index'] },
        'el-icon': { template: '<span><slot /></span>' },
      },
    },
  })
}

describe('AdminLayout', () => {
  it('renders a back-to-wiki link', async () => {
    const wrapper = await mountWithRouter()
    const link = wrapper.find('[data-test="back-to-wiki"]')
    expect(link.exists()).toBe(true)
    expect(link.text()).toContain('返回 Wiki')
  })
})
