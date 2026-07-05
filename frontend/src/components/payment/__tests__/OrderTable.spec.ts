import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import type { PaymentOrder } from '@/types/payment'
import OrderTable from '../OrderTable.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

const DataTableStub = {
  props: ['columns', 'data', 'loading'],
  template: `
    <div>
      <span v-for="column in columns" :key="column.key" data-test="column">{{ column.key }}</span>
    </div>
  `,
}

function orderFactory(overrides: Partial<PaymentOrder> = {}): PaymentOrder {
  return {
    id: 114,
    user_id: 10,
    amount: 99,
    pay_amount: 99,
    currency: 'CNY',
    fee_rate: 0,
    payment_type: 'wxpay',
    out_trade_no: 'sub2_20260701kIHSbcxB',
    status: 'COMPLETED',
    order_type: 'subscription',
    created_at: '2026-07-01T20:39:22Z',
    expires_at: '2026-07-01T21:09:22Z',
    refund_amount: 0,
    ...overrides,
  }
}

function renderColumnKeys(props: { showInternalId?: boolean } = {}): string[] {
  const wrapper = mount(OrderTable, {
    props: {
      orders: [orderFactory()],
      loading: false,
      ...props,
    },
    global: {
      stubs: {
        DataTable: DataTableStub,
        OrderStatusBadge: true,
      },
    },
  })

  return wrapper.findAll('[data-test="column"]').map(column => column.text())
}

describe('OrderTable', () => {
  it('hides the internal order id by default', () => {
    const columns = renderColumnKeys()

    expect(columns).not.toContain('id')
    expect(columns).toContain('out_trade_no')
  })

  it('can show the internal order id for admin views', () => {
    expect(renderColumnKeys({ showInternalId: true })).toEqual(expect.arrayContaining(['id', 'out_trade_no']))
  })
})
