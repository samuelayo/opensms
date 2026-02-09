import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseButton from './BaseButton.vue'

describe('BaseButton', () => {
  it('renders button with text', () => {
    const wrapper = mount(BaseButton, {
      slots: {
        default: 'Click me'
      }
    })

    expect(wrapper.text()).toBe('Click me')
    expect(wrapper.element.tagName).toBe('BUTTON')
  })

  it('applies primary variant class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        variant: 'primary'
      },
      slots: {
        default: 'Primary'
      }
    })

    expect(wrapper.classes()).toContain('btn-primary')
  })

  it('applies secondary variant class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        variant: 'secondary'
      },
      slots: {
        default: 'Secondary'
      }
    })

    expect(wrapper.classes()).toContain('btn-secondary')
  })

  it('applies danger variant class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        variant: 'danger'
      },
      slots: {
        default: 'Delete'
      }
    })

    expect(wrapper.classes()).toContain('btn-danger')
  })

  it('applies success variant class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        variant: 'success'
      },
      slots: {
        default: 'Save'
      }
    })

    expect(wrapper.classes()).toContain('btn-success')
  })

  it('handles click events', async () => {
    const wrapper = mount(BaseButton, {
      slots: {
        default: 'Click'
      }
    })

    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeTruthy()
    expect(wrapper.emitted('click')?.[0]).toBeTruthy()
  })

  it('disables button when disabled prop is true', () => {
    const wrapper = mount(BaseButton, {
      props: {
        disabled: true
      },
      slots: {
        default: 'Disabled'
      }
    })

    expect(wrapper.attributes('disabled')).toBeDefined()
    expect(wrapper.classes()).toContain('btn-disabled')
  })

  it('shows loading state', () => {
    const wrapper = mount(BaseButton, {
      props: {
        loading: true
      },
      slots: {
        default: 'Loading'
      }
    })

    expect(wrapper.classes()).toContain('btn-loading')
    expect(wrapper.attributes('disabled')).toBeDefined()
  })

  it('applies small size class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        size: 'sm'
      },
      slots: {
        default: 'Small'
      }
    })

    expect(wrapper.classes()).toContain('btn-sm')
  })

  it('applies large size class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        size: 'lg'
      },
      slots: {
        default: 'Large'
      }
    })

    expect(wrapper.classes()).toContain('btn-lg')
  })

  it('applies full width class', () => {
    const wrapper = mount(BaseButton, {
      props: {
        fullWidth: true
      },
      slots: {
        default: 'Full Width'
      }
    })

    expect(wrapper.classes()).toContain('w-full')
  })

  it('renders with icon slot', () => {
    const wrapper = mount(BaseButton, {
      slots: {
        default: 'With Icon',
        icon: '<svg>Icon</svg>'
      }
    })

    expect(wrapper.html()).toContain('<svg>Icon</svg>')
  })

  it('prevents click when disabled', async () => {
    const wrapper = mount(BaseButton, {
      props: {
        disabled: true
      },
      slots: {
        default: 'Disabled'
      }
    })

    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeFalsy()
  })

  it('prevents click when loading', async () => {
    const wrapper = mount(BaseButton, {
      props: {
        loading: true
      },
      slots: {
        default: 'Loading'
      }
    })

    await wrapper.trigger('click')
    expect(wrapper.emitted('click')).toBeFalsy()
  })

  it('applies custom classes', () => {
    const wrapper = mount(BaseButton, {
      props: {
        class: 'custom-class'
      },
      slots: {
        default: 'Custom'
      }
    })

    expect(wrapper.classes()).toContain('custom-class')
  })

  it('sets button type attribute', () => {
    const wrapper = mount(BaseButton, {
      props: {
        type: 'submit'
      },
      slots: {
        default: 'Submit'
      }
    })

    expect(wrapper.attributes('type')).toBe('submit')
  })
})
