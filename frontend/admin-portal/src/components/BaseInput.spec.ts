import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BaseInput from './BaseInput.vue'

describe('BaseInput', () => {
  it('renders input field', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        label: 'Email'
      }
    })

    expect(wrapper.find('input').exists()).toBe(true)
    expect(wrapper.find('label').text()).toBe('Email')
  })

  it('updates value on input', async () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        'onUpdate:modelValue': (e: string) => wrapper.setProps({ modelValue: e })
      }
    })

    const input = wrapper.find('input')
    await input.setValue('test@example.com')

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['test@example.com'])
  })

  it('displays error message', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        error: 'This field is required'
      }
    })

    expect(wrapper.text()).toContain('This field is required')
    expect(wrapper.classes()).toContain('input-error')
  })

  it('sets input type', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        type: 'password'
      }
    })

    expect(wrapper.find('input').attributes('type')).toBe('password')
  })

  it('disables input when disabled prop is true', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        disabled: true
      }
    })

    expect(wrapper.find('input').attributes('disabled')).toBeDefined()
  })

  it('sets placeholder text', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        placeholder: 'Enter your email'
      }
    })

    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter your email')
  })

  it('marks input as required', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        required: true
      }
    })

    expect(wrapper.find('input').attributes('required')).toBeDefined()
  })

  it('applies autocomplete attribute', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        autocomplete: 'email'
      }
    })

    expect(wrapper.find('input').attributes('autocomplete')).toBe('email')
  })

  it('shows help text', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        helpText: 'Enter a valid email address'
      }
    })

    expect(wrapper.text()).toContain('Enter a valid email address')
  })

  it('applies readonly attribute', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: 'readonly value',
        readonly: true
      }
    })

    expect(wrapper.find('input').attributes('readonly')).toBeDefined()
  })

  it('sets maxlength attribute', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        maxlength: 50
      }
    })

    expect(wrapper.find('input').attributes('maxlength')).toBe('50')
  })

  it('sets minlength attribute', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: '',
        minlength: 8
      }
    })

    expect(wrapper.find('input').attributes('minlength')).toBe('8')
  })

  it('renders with icon prefix', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: ''
      },
      slots: {
        prefix: '<svg>Icon</svg>'
      }
    })

    expect(wrapper.html()).toContain('<svg>Icon</svg>')
  })

  it('renders with icon suffix', () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: ''
      },
      slots: {
        suffix: '<svg>Icon</svg>'
      }
    })

    expect(wrapper.html()).toContain('<svg>Icon</svg>')
  })

  it('handles focus event', async () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: ''
      }
    })

    await wrapper.find('input').trigger('focus')
    expect(wrapper.emitted('focus')).toBeTruthy()
  })

  it('handles blur event', async () => {
    const wrapper = mount(BaseInput, {
      props: {
        modelValue: ''
      }
    })

    await wrapper.find('input').trigger('blur')
    expect(wrapper.emitted('blur')).toBeTruthy()
  })
})
