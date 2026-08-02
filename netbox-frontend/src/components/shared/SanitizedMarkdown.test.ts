import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SanitizedMarkdown from './SanitizedMarkdown.vue'

describe('SanitizedMarkdown', () => {
  it('escapes raw HTML and event handlers', () => {
    const wrapper = mount(SanitizedMarkdown, {
      props: { content: '<img src=x onerror="alert(1)"><script>alert(2)</script>' },
    })

    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.find('script').exists()).toBe(false)
    expect(wrapper.html()).toContain('&lt;img')
    expect(wrapper.html()).toContain('&lt;script&gt;')
  })

  it('does not turn unsafe protocols into links', () => {
    const wrapper = mount(SanitizedMarkdown, {
      props: { content: '[click me](javascript:alert(1))' },
    })

    expect(wrapper.find('a').exists()).toBe(false)
    expect(wrapper.text()).toContain('javascript:alert(1)')
  })

  it('allows HTTPS links with safe external-link attributes', () => {
    const wrapper = mount(SanitizedMarkdown, {
      props: { content: '[NetBox](https://netboxlabs.com/)' },
    })
    const link = wrapper.get('a')

    expect(link.attributes('href')).toBe('https://netboxlabs.com/')
    expect(link.attributes('target')).toBe('_blank')
    expect(link.attributes('rel')).toBe('noopener noreferrer')
  })
})
