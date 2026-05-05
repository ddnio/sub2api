<template>
  <AppLayout>
<<<<<<< HEAD
    <div class="mx-auto max-w-4xl space-y-6">
      <div class="grid grid-cols-1 gap-6 sm:grid-cols-3">
        <StatCard :title="t('profile.accountBalance')" :value="formatCurrency(user?.balance || 0)" :icon="WalletIcon" icon-variant="success" />
        <StatCard :title="t('profile.concurrencyLimit')" :value="user?.concurrency || 0" :icon="BoltIcon" icon-variant="warning" />
        <StatCard :title="t('profile.memberSince')" :value="formatDate(user?.created_at || '', { year: 'numeric', month: 'long' })" :icon="CalendarIcon" icon-variant="primary" />
      </div>
      <ProfileInfoCard :user="profileUser || user" />
      <div v-if="contactInfo" class="card border-primary-200 bg-primary-50 dark:bg-primary-900/20 p-6">
=======
    <div
      data-testid="profile-shell"
      class="mx-auto max-w-[950px] space-y-6"
    >
      <ProfileInfoCard
        :user="user"
        :linuxdo-enabled="linuxdoOAuthEnabled"
        :oidc-enabled="oidcOAuthEnabled"
        :oidc-provider-name="oidcOAuthProviderName"
        :wechat-enabled="wechatOAuthEnabled"
        :wechat-open-enabled="wechatOAuthOpenEnabled"
        :wechat-mp-enabled="wechatOAuthMPEnabled"
      />

      <div
        v-if="contactInfo"
        class="card border-primary-200 bg-primary-50 p-6 dark:bg-primary-900/20"
      >
>>>>>>> v0.1.116
        <div class="flex items-center gap-4">
          <div class="rounded-xl bg-primary-100 p-3 text-primary-600">
            <Icon name="chat" size="lg" />
          </div>
          <div>
            <h3 class="font-semibold text-primary-800 dark:text-primary-200">
              {{ t('common.contactSupport') }}
            </h3>
            <p class="text-sm font-medium">{{ contactInfo }}</p>
          </div>
        </div>
      </div>
<<<<<<< HEAD
      <ProfileEditForm :initial-username="profileUser?.username || user?.username || ''" @updated="handleProfileUpdated" />
      <ProfileBalanceNotifyCard
        v-if="profileUser"
        :enabled="profileUser.balance_notify_enabled"
        :threshold="profileUser.balance_notify_threshold"
        :extra-emails="profileUser.balance_notify_extra_emails || []"
        :system-default-threshold="publicSettings?.balance_low_notify_threshold || 0"
        :user-email="profileUser.email"
      />
      <ProfileIdentityBindingsSection
        :user="profileUser"
        :linuxdo-enabled="publicSettings?.linuxdo_oauth_enabled || false"
        :wechat-enabled="wechatOAuthEnabledForCurrentBrowser"
        :wechat-open-enabled="publicSettings?.wechat_oauth_open_enabled"
        :wechat-mp-enabled="publicSettings?.wechat_oauth_mp_enabled"
        :oidc-enabled="publicSettings?.oidc_oauth_enabled || false"
        :oidc-provider-name="publicSettings?.oidc_oauth_provider_name || 'OIDC'"
        @updated="handleProfileUpdated"
      />
=======

>>>>>>> v0.1.116
      <ProfilePasswordForm />

      <ProfileBalanceNotifyCard
        v-if="user && balanceLowNotifyEnabled"
        :enabled="user.balance_notify_enabled ?? true"
        :threshold="user.balance_notify_threshold"
        :extra-emails="user.balance_notify_extra_emails ?? []"
        :system-default-threshold="systemDefaultThreshold"
        :user-email="user.email"
      />

      <ProfileTotpCard />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
<<<<<<< HEAD
import { ref, computed, h, onMounted } from 'vue'; import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'; import { formatDate } from '@/utils/format'
import { authAPI, userAPI } from '@/api'; import AppLayout from '@/components/layout/AppLayout.vue'
import StatCard from '@/components/common/StatCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfileEditForm from '@/components/user/profile/ProfileEditForm.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileIdentityBindingsSection from '@/components/user/profile/ProfileIdentityBindingsSection.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import { Icon } from '@/components/icons'
import type { PublicSettings, UserProfile } from '@/types'
=======
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@/components/icons'
import AppLayout from '@/components/layout/AppLayout.vue'
import ProfileBalanceNotifyCard from '@/components/user/profile/ProfileBalanceNotifyCard.vue'
import ProfileInfoCard from '@/components/user/profile/ProfileInfoCard.vue'
import ProfilePasswordForm from '@/components/user/profile/ProfilePasswordForm.vue'
import ProfileTotpCard from '@/components/user/profile/ProfileTotpCard.vue'
import { isWeChatWebOAuthEnabled } from '@/api/auth'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const user = computed(() => authStore.user)
>>>>>>> v0.1.116

const contactInfo = ref('')
<<<<<<< HEAD
const publicSettings = ref<PublicSettings | null>(null)
const profileUser = ref<UserProfile | null>(null)
const isWeChatBrowser = computed(() => /MicroMessenger/i.test(navigator.userAgent))
const wechatOAuthEnabledForCurrentBrowser = computed(() => {
  const settings = publicSettings.value
  if (!settings?.wechat_oauth_enabled) return false
  return isWeChatBrowser.value
    ? settings.wechat_oauth_mp_enabled
    : settings.wechat_oauth_open_enabled
})
=======
const balanceLowNotifyEnabled = ref(false)
const systemDefaultThreshold = ref(0)
const linuxdoOAuthEnabled = ref(false)
const wechatOAuthEnabled = ref(false)
const wechatOAuthOpenEnabled = ref<boolean | undefined>(undefined)
const wechatOAuthMPEnabled = ref<boolean | undefined>(undefined)
const oidcOAuthEnabled = ref(false)
const oidcOAuthProviderName = ref('OIDC')
>>>>>>> v0.1.116

onMounted(async () => {
  const profileRefresh = authStore.refreshUser().catch((error) => {
    console.error('Failed to refresh profile:', error)
  })

<<<<<<< HEAD
onMounted(async () => {
  try {
    const s = await authAPI.getPublicSettings()
    publicSettings.value = s
    contactInfo.value = s.contact_info || ''
  } catch (error) {
    console.error('Failed to load contact info:', error)
  }
  try {
    profileUser.value = await userAPI.getProfile()
  } catch (error) {
    console.error('Failed to load profile:', error)
  }
})

const handleProfileUpdated = (updated: UserProfile) => {
  profileUser.value = updated
}
const formatCurrency = (v: number) => `$${v.toFixed(2)}`
=======
  const settingsLoad = appStore.fetchPublicSettings()
    .then((settings) => {
      if (!settings) {
        return
      }
      contactInfo.value = settings.contact_info || ''
      balanceLowNotifyEnabled.value = settings.balance_low_notify_enabled ?? false
      systemDefaultThreshold.value = settings.balance_low_notify_threshold ?? 0
      linuxdoOAuthEnabled.value = settings.linuxdo_oauth_enabled ?? false
      wechatOAuthEnabled.value = isWeChatWebOAuthEnabled(settings)
      wechatOAuthOpenEnabled.value = typeof settings.wechat_oauth_open_enabled === 'boolean'
        ? settings.wechat_oauth_open_enabled
        : undefined
      wechatOAuthMPEnabled.value = typeof settings.wechat_oauth_mp_enabled === 'boolean'
        ? settings.wechat_oauth_mp_enabled
        : undefined
      oidcOAuthEnabled.value = settings.oidc_oauth_enabled ?? false
      oidcOAuthProviderName.value = settings.oidc_oauth_provider_name || 'OIDC'
    })
    .catch((error) => {
      console.error('Failed to load settings:', error)
    })

  await Promise.all([profileRefresh, settingsLoad])
})
>>>>>>> v0.1.116
</script>
