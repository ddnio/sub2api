<template>
  <div class="card p-6">
    <div class="mb-4">
      <h3 class="text-base font-semibold text-gray-900 dark:text-white">
        {{ t('profile.authBindings.title') }}
      </h3>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('profile.authBindings.description') }}
      </p>
    </div>

    <div class="space-y-3">
      <div
        v-for="item in providerItems"
        :key="item.provider"
        class="flex items-center justify-between gap-3 rounded-lg border border-gray-100 px-4 py-3 dark:border-dark-700"
      >
        <div class="min-w-0">
          <div class="text-sm font-medium text-gray-900 dark:text-white">
            {{ item.label }}
          </div>
          <div v-if="item.hint" class="mt-1 truncate text-xs text-gray-500 dark:text-gray-400">
            {{ item.hint }}
          </div>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <span
            :data-testid="`profile-binding-${item.provider}-status`"
            :class="['badge', item.bound ? 'badge-success' : 'badge-gray']"
          >
            {{
              item.bound
                ? t('profile.authBindings.status.bound')
                : t('profile.authBindings.status.notBound')
            }}
          </span>

          <button
            v-if="item.canBind"
            :data-testid="`profile-binding-${item.provider}-action`"
            type="button"
            class="btn btn-secondary btn-sm"
            @click="handleBind(item.provider)"
          >
            {{ t('profile.authBindings.bindAction', { providerName: item.label }) }}
          </button>
          <button
            v-if="item.canUnbind"
            :data-testid="`profile-binding-${item.provider}-unbind`"
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="unbindingProvider === item.provider"
            @click="handleUnbind(item.provider)"
          >
            {{ t('profile.authBindings.unbindAction', { providerName: item.label }) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import {
  hasExplicitWeChatOAuthCapabilities,
  resolveWeChatOAuthStartStrict,
  type WeChatOAuthPublicSettings,
} from '@/api/auth'
import { startOAuthBinding, unbindAuthIdentity as unbindAuthProvider } from '@/api/user'
import { useAppStore } from '@/stores'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfile } from '@/types'

const props = withDefaults(
  defineProps<{
    user: UserProfile | User | null
    linuxdoEnabled?: boolean
    wechatEnabled?: boolean
    wechatOpenEnabled?: boolean
    wechatMpEnabled?: boolean
    oidcEnabled?: boolean
    oidcProviderName?: string
  }>(),
  {
    linuxdoEnabled: false,
    wechatEnabled: false,
    wechatOpenEnabled: undefined,
    wechatMpEnabled: undefined,
    oidcEnabled: false,
    oidcProviderName: 'OIDC',
  }
)

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const unbindingProvider = ref<UserAuthProvider | null>(null)

const wechatOAuthSettings = computed<WeChatOAuthPublicSettings | null>(() => {
  if (hasExplicitWeChatOAuthCapabilities(appStore.cachedPublicSettings)) {
    return appStore.cachedPublicSettings
  }

  if (typeof props.wechatOpenEnabled === 'boolean' && typeof props.wechatMpEnabled === 'boolean') {
    return {
      wechat_oauth_enabled: props.wechatEnabled,
      wechat_oauth_open_enabled: props.wechatOpenEnabled,
      wechat_oauth_mp_enabled: props.wechatMpEnabled,
    }
  }

  return null
})

const resolvedWeChatBinding = computed(() => resolveWeChatOAuthStartStrict(wechatOAuthSettings.value))

type BindingLike = boolean | UserAuthBindingStatus | undefined

const emit = defineEmits<{
  updated: [user: User]
}>()

function normalizeBindingStatus(
  provider: UserAuthProvider,
  binding: BindingLike
): UserAuthBindingStatus | undefined {
  if (typeof binding === 'boolean') {
    return {
      provider,
      bound: binding
    }
  }
  return binding
}

function getStatus(provider: UserAuthProvider): UserAuthBindingStatus | undefined {
  return (
    props.user?.identities?.[provider] ??
    normalizeBindingStatus(provider, props.user?.auth_bindings?.[provider]) ??
    normalizeBindingStatus(provider, props.user?.identity_bindings?.[provider])
  )
}

const emailStatus = computed(() => getStatus('email'))

const emailBound = computed(() => {
  if (typeof props.user?.email_bound === 'boolean') {
    return props.user.email_bound
  }
  const status = emailStatus.value
  if (typeof status?.bound === 'boolean') {
    return status.bound
  }
  const email = props.user?.email?.trim() || ''
  return Boolean(email && !email.endsWith('.invalid'))
})

const boundEmail = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) return ''
  if (email.endsWith('.invalid') && !emailBound.value) return ''
  return email
})

function getEmailStatus(): UserAuthBindingStatus | undefined {
  const status = emailStatus.value
  if (status) {
    return {
      ...status,
      bound: emailBound.value
    }
  }
  return {
    provider: 'email',
    bound: emailBound.value
  }
}

function providerHint(status: UserAuthBindingStatus | undefined): string {
  return status?.display_name || status?.subject_hint || status?.note || ''
}

const providerItems = computed(() => [
  {
    provider: 'email' as const,
    label: t('profile.authBindings.providers.email'),
    bound: emailBound.value,
    canBind: false,
    canUnbind: false,
    hint: providerHint(getEmailStatus()) || boundEmail.value,
  },
  {
    provider: 'linuxdo' as const,
    label: t('profile.authBindings.providers.linuxdo'),
    bound: Boolean(getStatus('linuxdo')?.bound),
    canBind: props.linuxdoEnabled && Boolean(getStatus('linuxdo')?.can_bind),
    canUnbind: Boolean(getStatus('linuxdo')?.can_unbind),
    hint: providerHint(getStatus('linuxdo')),
  },
  {
    provider: 'wechat' as const,
    label: t('profile.authBindings.providers.wechat'),
    bound: Boolean(getStatus('wechat')?.bound),
    canBind: resolvedWeChatBinding.value.mode !== null && Boolean(getStatus('wechat')?.can_bind),
    canUnbind: Boolean(getStatus('wechat')?.can_unbind),
    hint: providerHint(getStatus('wechat')),
  },
  {
    provider: 'oidc' as const,
    label: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
    bound: Boolean(getStatus('oidc')?.bound),
    canBind: props.oidcEnabled && Boolean(getStatus('oidc')?.can_bind),
    canUnbind: Boolean(getStatus('oidc')?.can_unbind),
    hint: providerHint(getStatus('oidc')),
  },
])

function handleBind(provider: UserAuthProvider): void {
  if (provider === 'email') {
    return
  }
  void startOAuthBinding(provider, {
    redirectTo: route.fullPath || '/profile',
    wechatOAuthSettings: provider === 'wechat' ? wechatOAuthSettings.value : null,
  })
}

async function handleUnbind(provider: UserAuthProvider): Promise<void> {
  if (provider === 'email') {
    return
  }
  unbindingProvider.value = provider
  try {
    const updated = await unbindAuthProvider(provider)
    emit('updated', updated)
  } finally {
    unbindingProvider.value = null
  }
}
</script>
