<template>
  <div class="space-y-6">
    <section
      data-testid="profile-overview-hero"
      class="card overflow-hidden border border-primary-100/80 bg-gradient-to-br from-primary-50 via-white to-amber-50/70 dark:border-primary-900/40 dark:from-primary-950/40 dark:via-dark-900 dark:to-dark-950"
    >
<<<<<<< HEAD
      <div class="flex items-center gap-4">
        <div
          class="flex h-16 w-16 items-center justify-center overflow-hidden rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-2xl font-bold text-white shadow-lg shadow-primary-500/20"
        >
          <img
            v-if="avatarUrl"
            :src="avatarUrl"
            :alt="displayName"
            class="h-full w-full object-cover"
          >
          <span v-else>{{ avatarInitial }}</span>
        </div>
        <div class="min-w-0 flex-1">
          <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">
            {{ user?.email }}
          </h2>
          <div class="mt-1 flex items-center gap-2">
            <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
              {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
            </span>
            <span
              :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']"
=======
      <div class="px-6 py-6 md:px-8">
        <div class="flex flex-col gap-6 lg:flex-row lg:items-start">
          <div
            class="flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-[1.75rem] bg-gradient-to-br from-primary-500 to-primary-600 text-2xl font-bold text-white shadow-lg shadow-primary-500/20"
          >
            <img
              v-if="avatarUrl"
              :src="avatarUrl"
              :alt="displayName"
              class="h-full w-full object-cover"
>>>>>>> v0.1.116
            >
            <span v-else>{{ avatarInitial }}</span>
          </div>

          <div class="min-w-0 flex-1 space-y-5">
            <div class="space-y-3">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="truncate text-2xl font-semibold text-gray-900 dark:text-white">
                  {{ displayName }}
                </h2>
                <span :class="['badge', user?.role === 'admin' ? 'badge-primary' : 'badge-gray']">
                  {{ user?.role === 'admin' ? t('profile.administrator') : t('profile.user') }}
                </span>
                <span
                  :class="['badge', user?.status === 'active' ? 'badge-success' : 'badge-danger']"
                >
                  {{
                    user?.status === 'active'
                      ? t('common.active')
                      : t('common.disabled')
                  }}
                </span>
              </div>

              <div class="space-y-1">
                <p class="truncate text-sm text-gray-600 dark:text-gray-300">
                  {{ primaryEmailDisplay }}
                </p>
                <div
                  v-if="sourceHints.length"
                  class="flex flex-wrap gap-2 text-xs text-gray-500 dark:text-gray-400"
                >
                  <span
                    v-for="hint in sourceHints"
                    :key="hint.key"
                    class="inline-flex items-center gap-1 rounded-full bg-white/80 px-3 py-1 ring-1 ring-primary-100 dark:bg-dark-900/70 dark:ring-primary-900/40"
                  >
                    <Icon name="link" size="sm" />
                    {{ hint.text }}
                  </span>
                </div>
              </div>
            </div>

            <div class="grid gap-3 sm:grid-cols-3">
              <div
                data-testid="profile-overview-metric-balance"
                class="rounded-2xl bg-white/85 px-4 py-3 shadow-sm ring-1 ring-white/70 dark:bg-dark-900/60 dark:ring-dark-700"
              >
                <p class="text-xs font-medium uppercase tracking-[0.16em] text-gray-400 dark:text-gray-500">
                  {{ t('profile.accountBalance') }}
                </p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                  {{ formatCurrency(user?.balance || 0) }}
                </p>
              </div>
              <div
                data-testid="profile-overview-metric-concurrency"
                class="rounded-2xl bg-white/85 px-4 py-3 shadow-sm ring-1 ring-white/70 dark:bg-dark-900/60 dark:ring-dark-700"
              >
                <p class="text-xs font-medium uppercase tracking-[0.16em] text-gray-400 dark:text-gray-500">
                  {{ t('profile.concurrencyLimit') }}
                </p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                  {{ user?.concurrency || 0 }}
                </p>
              </div>
              <div
                data-testid="profile-overview-metric-member-since"
                class="rounded-2xl bg-white/85 px-4 py-3 shadow-sm ring-1 ring-white/70 dark:bg-dark-900/60 dark:ring-dark-700"
              >
                <p class="text-xs font-medium uppercase tracking-[0.16em] text-gray-400 dark:text-gray-500">
                  {{ t('profile.memberSince') }}
                </p>
                <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                  {{ memberSinceLabel }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div class="space-y-6">
      <div data-testid="profile-main-column" class="space-y-6">
        <section
          data-testid="profile-basics-panel"
          class="card border border-gray-100 bg-white/90 p-6 dark:border-dark-700 dark:bg-dark-900/50"
        >
          <div class="mb-5 flex items-start justify-between gap-4">
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('profile.basicsTitle') }}
              </h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {{ t('profile.basicsDescription') }}
              </p>
            </div>
          </div>

          <div class="grid gap-6 sm:grid-cols-1 md:grid-cols-2">
            <div class="rounded-3xl border border-gray-100 bg-gray-50/80 p-5 dark:border-dark-700 dark:bg-dark-900/30">
              <ProfileAvatarCard
                :user="user"
                embedded
              />
            </div>

            <div class="rounded-3xl border border-gray-100 bg-gray-50/80 p-5 dark:border-dark-700 dark:bg-dark-900/30">
              <ProfileEditForm
                :initial-username="user?.username || ''"
                embedded
              />
            </div>
          </div>
        </section>

        <section
          data-testid="profile-auth-bindings-panel"
          class="card border border-gray-100 bg-white/90 p-6 dark:border-dark-700 dark:bg-dark-900/50"
        >
          <ProfileIdentityBindingsSection
            :user="user"
            :linuxdo-enabled="linuxdoEnabled"
            :oidc-enabled="oidcEnabled"
            :oidc-provider-name="oidcProviderName"
            :wechat-enabled="wechatEnabled"
            :wechat-open-enabled="wechatOpenEnabled"
            :wechat-mp-enabled="wechatMpEnabled"
            embedded
            compact
          />
        </section>
      </div>

      <div data-testid="profile-side-column" class="space-y-6">
        <section
          v-if="sourceHints.length"
          class="card border border-gray-100 bg-white/90 p-6 dark:border-dark-700 dark:bg-dark-900/50"
        >
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('profile.linkedProfileSources') }}
          </h3>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('profile.linkedProfileSourcesDescription') }}
          </p>

          <div class="mt-5 grid gap-3">
            <div
              v-for="hint in sourceHints"
              :key="hint.key"
              class="flex items-start gap-3 rounded-2xl border border-gray-100 bg-gray-50/80 px-4 py-3 text-sm text-gray-600 dark:border-dark-700 dark:bg-dark-900/30 dark:text-gray-300"
            >
              <Icon name="link" size="sm" class="mt-0.5 text-gray-400 dark:text-gray-500" />
              <span>{{ hint.text }}</span>
            </div>
          </div>
        </section>
      </div>

      <div
        v-if="sourceHints.length"
        class="mt-4 grid gap-2 rounded-2xl border border-gray-100 bg-gray-50/80 p-3 text-xs text-gray-500 dark:border-dark-700 dark:bg-dark-900/30 dark:text-gray-400"
      >
        <div
          v-for="hint in sourceHints"
          :key="hint.key"
          class="flex items-start gap-2"
        >
          <Icon name="link" size="sm" class="mt-0.5 text-gray-400 dark:text-gray-500" />
          <span>{{ hint.text }}</span>
        </div>
      </div>

      <div
        class="mt-4 rounded-2xl border border-gray-100 bg-white/90 p-4 dark:border-dark-700 dark:bg-dark-900/30"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('profile.avatar.title') }}
            </h3>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('profile.avatar.description') }}
            </p>
          </div>
          <button
            data-testid="profile-avatar-delete"
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="avatarSaving"
            @click="handleAvatarDelete"
          >
            {{ t('common.delete') }}
          </button>
        </div>

        <div class="mt-3 space-y-3">
          <label
            for="profile-avatar-input"
            class="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400"
          >
            {{ t('profile.avatar.inputLabel') }}
          </label>
          <textarea
            id="profile-avatar-input"
            data-testid="profile-avatar-input"
            v-model="avatarDraft"
            rows="3"
            class="input min-h-[88px]"
            :placeholder="t('profile.avatar.inputPlaceholder')"
          />
          <div class="flex flex-wrap items-center gap-2">
            <label class="btn btn-secondary btn-sm cursor-pointer">
              <input
                data-testid="profile-avatar-file-input"
                type="file"
                accept="image/*"
                class="hidden"
                @change="handleAvatarFileChange"
              >
              {{ t('profile.avatar.uploadAction') }}
            </label>
            <button
              data-testid="profile-avatar-save"
              type="button"
              class="btn btn-primary btn-sm"
              :disabled="avatarSaving"
              @click="handleAvatarSave"
            >
              {{ t('common.save') }}
            </button>
            <span class="text-xs text-gray-400 dark:text-gray-500">
              {{ t('profile.avatar.uploadHint') }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
<<<<<<< HEAD
import { computed, ref, watch } from 'vue'
=======
import { computed } from 'vue'
>>>>>>> v0.1.116
import { useI18n } from 'vue-i18n'
import { userAPI } from '@/api'
import Icon from '@/components/icons/Icon.vue'
<<<<<<< HEAD
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import type { User, UserAuthProvider, UserProfile, UserProfileSourceContext } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'

const props = withDefaults(
  defineProps<{
    user: UserProfile | User | null
  }>(),
  {}
)

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const maxAvatarBytes = 100 * 1024
const targetAvatarUploadBytes = 20 * 1024
const avatarScaleSteps = [1, 0.92, 0.84, 0.76, 0.68, 0.6, 0.52, 0.44, 0.36]
const avatarQualitySteps = [0.92, 0.84, 0.76, 0.68, 0.6, 0.52, 0.44, 0.36]
const avatarDraft = ref(props.user?.avatar_url?.trim() || '')
const avatarSaving = ref(false)
=======
import ProfileAvatarCard from '@/components/user/profile/ProfileAvatarCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import type { User, UserAuthBindingStatus, UserAuthProvider, UserProfileSourceContext } from '@/types'

const props = withDefaults(defineProps<{
  user: User | null
  linuxdoEnabled?: boolean
  oidcEnabled?: boolean
  oidcProviderName?: string
  wechatEnabled?: boolean
  wechatOpenEnabled?: boolean
  wechatMpEnabled?: boolean
}>(), {
  linuxdoEnabled: false,
  oidcEnabled: false,
  oidcProviderName: 'OIDC',
  wechatEnabled: false,
  wechatOpenEnabled: undefined,
  wechatMpEnabled: undefined,
})

const { t } = useI18n()

function normalizeBindingStatus(binding: boolean | UserAuthBindingStatus | undefined): boolean | null {
  if (typeof binding === 'boolean') {
    return binding
  }
  if (!binding) {
    return null
  }
  if (typeof binding.bound === 'boolean') {
    return binding.bound
  }
  return Boolean(binding.provider_subject || binding.issuer || binding.provider_key)
}

function isEmailBound(user: User | null | undefined): boolean {
  if (typeof user?.email_bound === 'boolean') {
    return user.email_bound
  }

  const nested = user?.auth_bindings?.email ?? user?.identity_bindings?.email
  const normalized = normalizeBindingStatus(nested)
  return normalized ?? false
}

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || t('profile.user'))
const primaryEmailDisplay = computed(() => {
  const email = props.user?.email?.trim() || ''
  if (!email) {
    return ''
  }
  if (email.endsWith('.invalid') && !isEmailBound(props.user)) {
    return ''
  }
  return email
})
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')
const memberSinceLabel = computed(() => {
  const raw = props.user?.created_at?.trim()
  if (!raw) {
    return '-'
  }

  const date = new Date(raw)
  if (Number.isNaN(date.getTime())) {
    return '-'
  }

  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
  }).format(date)
})
>>>>>>> v0.1.116

const providerLabels = computed<Record<UserAuthProvider, string>>(() => ({
  email: t('profile.authBindings.providers.email'),
  linuxdo: t('profile.authBindings.providers.linuxdo'),
<<<<<<< HEAD
  oidc: t('profile.authBindings.providers.oidc', { providerName: 'OIDC' }),
  wechat: t('profile.authBindings.providers.wechat'),
}))

const avatarUrl = computed(() => props.user?.avatar_url?.trim() || '')
const displayName = computed(() => props.user?.username?.trim() || props.user?.email?.trim() || 'User')
const avatarInitial = computed(() => displayName.value.charAt(0).toUpperCase() || 'U')

watch(
  () => props.user?.avatar_url,
  (value) => {
    avatarDraft.value = value?.trim() || ''
  }
)
=======
  oidc: t('profile.authBindings.providers.oidc', { providerName: props.oidcProviderName }),
  wechat: t('profile.authBindings.providers.wechat')
}))

function formatCurrency(value: number): string {
  return `$${value.toFixed(2)}`
}
>>>>>>> v0.1.116

function normalizeProvider(value: string): UserAuthProvider | null {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'email' || normalized === 'linuxdo' || normalized === 'wechat') {
    return normalized
  }
  if (normalized === 'oidc' || normalized.startsWith('oidc:') || normalized.startsWith('oidc/')) {
    return 'oidc'
  }
  return null
}

function readObjectString(source: Record<string, unknown>, ...keys: string[]): string {
  for (const key of keys) {
    const value = source[key]
    if (typeof value === 'string' && value.trim()) {
      return value.trim()
    }
  }
  return ''
}

function resolveThirdPartySource(
  rawSource: string | UserProfileSourceContext | null | undefined
): { provider: UserAuthProvider; label: string } | null {
  if (!rawSource) {
    return null
  }

  if (typeof rawSource === 'string') {
    const provider = normalizeProvider(rawSource)
    if (!provider || provider === 'email') {
      return null
    }
    return {
      provider,
<<<<<<< HEAD
      label: providerLabels.value[provider],
=======
      label: providerLabels.value[provider]
>>>>>>> v0.1.116
    }
  }

  const sourceRecord = rawSource as Record<string, unknown>
  const provider = normalizeProvider(
    readObjectString(sourceRecord, 'provider', 'source', 'provider_type', 'auth_provider')
  )
  if (!provider || provider === 'email') {
    return null
  }

  const explicitLabel = readObjectString(
    sourceRecord,
    'provider_label',
    'label',
    'provider_name',
    'providerName'
  )

  return {
    provider,
<<<<<<< HEAD
    label: explicitLabel || providerLabels.value[provider],
=======
    label: explicitLabel || providerLabels.value[provider]
>>>>>>> v0.1.116
  }
}

const sourceHints = computed(() => {
  const currentUser = props.user
  if (!currentUser) {
    return []
  }

  const hints: Array<{ key: string; text: string }> = []
  const avatarSource = resolveThirdPartySource(
    currentUser.profile_sources?.avatar ?? currentUser.avatar_source
  )
  const usernameSource = resolveThirdPartySource(
    currentUser.profile_sources?.username ??
      currentUser.profile_sources?.display_name ??
      currentUser.profile_sources?.nickname ??
      currentUser.display_name_source ??
      currentUser.username_source ??
      currentUser.nickname_source
  )

  if (avatarSource) {
    hints.push({
      key: 'avatar',
<<<<<<< HEAD
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label }),
=======
      text: t('profile.authBindings.source.avatar', { providerName: avatarSource.label })
>>>>>>> v0.1.116
    })
  }

  if (usernameSource) {
    hints.push({
      key: 'username',
<<<<<<< HEAD
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label }),
=======
      text: t('profile.authBindings.source.username', { providerName: usernameSource.label })
>>>>>>> v0.1.116
    })
  }

  return hints
})
<<<<<<< HEAD

function estimateDataURLByteSize(value: string): number {
  const [, encoded = ''] = value.split(',', 2)
  const sanitized = encoded.replace(/\s+/g, '')
  const padding = sanitized.endsWith('==') ? 2 : sanitized.endsWith('=') ? 1 : 0
  return Math.max(0, Math.floor((sanitized.length * 3) / 4) - padding)
}

function validateAvatarInput(value: string): string | null {
  const normalized = value.trim()
  if (!normalized) {
    return null
  }

  if (normalized.startsWith('data:')) {
    if (!/^data:image\/[a-zA-Z0-9.+-]+;base64,/i.test(normalized)) {
      appStore.showError(t('profile.avatar.invalidValue'))
      return null
    }
    if (estimateDataURLByteSize(normalized) > maxAvatarBytes) {
      appStore.showError(t('profile.avatar.fileTooLarge'))
      return null
    }
    return normalized
  }

  try {
    const parsed = new URL(normalized)
    if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
      return normalized
    }
  } catch {
    // Invalid URL is handled below.
  }

  appStore.showError(t('profile.avatar.invalidValue'))
  return null
}

function readFileAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(typeof reader.result === 'string' ? reader.result : '')
    reader.onerror = () => reject(reader.error ?? new Error('avatar_read_failed'))
    reader.readAsDataURL(file)
  })
}

function loadImage(dataURL: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error(t('profile.avatar.readFailed')))
    image.src = dataURL
  })
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (!blob) {
        reject(new Error(t('profile.avatar.compressFailed')))
        return
      }
      resolve(blob)
    }, type, quality)
  })
}

async function compressAvatarFile(file: File): Promise<File> {
  const sourceDataURL = await readFileAsDataURL(file)
  const image = await loadImage(sourceDataURL)
  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error(t('profile.avatar.compressFailed'))
  }

  for (const scale of avatarScaleSteps) {
    const width = Math.max(1, Math.round(image.naturalWidth * scale))
    const height = Math.max(1, Math.round(image.naturalHeight * scale))
    canvas.width = width
    canvas.height = height
    ctx.clearRect(0, 0, width, height)
    ctx.drawImage(image, 0, 0, width, height)

    for (const quality of avatarQualitySteps) {
      const blob = await canvasToBlob(canvas, 'image/webp', quality)
      if (blob.size <= targetAvatarUploadBytes) {
        const fileName = file.name.replace(/\.[^.]+$/, '') || 'avatar'
        return new File([blob], `${fileName}.webp`, { type: 'image/webp' })
      }
    }
  }

  throw new Error(t('profile.avatar.compressTooLarge'))
}

async function prepareAvatarUpload(file: File): Promise<File> {
  if (!file.type.startsWith('image/')) {
    throw new Error(t('profile.avatar.invalidType'))
  }
  if (file.type === 'image/gif') {
    if (file.size > targetAvatarUploadBytes) {
      throw new Error(t('profile.avatar.gifTooLarge'))
    }
    return file
  }
  if (file.size <= targetAvatarUploadBytes) {
    return file
  }
  return compressAvatarFile(file)
}

async function handleAvatarFileChange(event: Event) {
  const input = event.target as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) {
    input.value = ''
  }
  if (!file) {
    return
  }
  if (!file.type.startsWith('image/')) {
    appStore.showError(t('profile.avatar.invalidType'))
    return
  }
  if (file.size > maxAvatarBytes) {
    appStore.showError(t('profile.avatar.fileTooLarge'))
    return
  }

  try {
    const preparedFile = await prepareAvatarUpload(file)
    const dataURL = await readFileAsDataURL(preparedFile)
    const normalized = validateAvatarInput(dataURL)
    if (!normalized) {
      return
    }
    avatarDraft.value = normalized
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  }
}

async function handleAvatarSave() {
  const normalized = validateAvatarInput(avatarDraft.value)
  if (!normalized) {
    return
  }

  avatarSaving.value = true
  try {
    const updated = await userAPI.updateProfile({ avatar_url: normalized })
    authStore.user = updated
    avatarDraft.value = updated.avatar_url?.trim() || ''
    appStore.showSuccess(t('profile.avatar.saveSuccess'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    avatarSaving.value = false
  }
}

async function handleAvatarDelete() {
  if (avatarSaving.value) {
    return
  }
  if (!avatarDraft.value.trim() && !props.user?.avatar_url?.trim()) {
    appStore.showError(t('profile.avatar.emptyDeleteHint'))
    return
  }

  avatarSaving.value = true
  try {
    const updated = await userAPI.updateProfile({ avatar_url: '' })
    authStore.user = updated
    avatarDraft.value = ''
    appStore.showSuccess(t('profile.avatar.deleteSuccess'))
  } catch (error: unknown) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    avatarSaving.value = false
  }
}
=======
>>>>>>> v0.1.116
</script>
