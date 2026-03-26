import { ref, watch, onMounted, onUnmounted, type Ref } from 'vue'

interface SyncedTabOptions {
  /** sync-group name, e.g. 'shell', 'os', 'lang' */
  group: string
  /** Page-level namespace to isolate from other pages, e.g. 'docs', 'usekey' */
  scope: string
  /** Available tab IDs for this specific code block */
  availableTabs: string[]
  /** Default tab if stored value is invalid or absent */
  defaultTab: string
}

function resolveStoredTab(stored: string | null, availableTabs: string[], defaultTab: string): string {
  if (stored && availableTabs.includes(stored)) return stored
  return defaultTab
}

export function useSyncedTabState(options: SyncedTabOptions): { activeTab: Ref<string> } {
  const storageKey = `${options.scope}:${options.group}`
  const stored = localStorage.getItem(storageKey)
  const initial = resolveStoredTab(stored, options.availableTabs, options.defaultTab)
  const activeTab = ref(initial)

  watch(activeTab, (val) => {
    localStorage.setItem(storageKey, val)
    window.dispatchEvent(new StorageEvent('storage', { key: storageKey, newValue: val }))
  })

  function onStorage(e: StorageEvent) {
    if (e.key !== storageKey || !e.newValue) return
    const resolved = resolveStoredTab(e.newValue, options.availableTabs, options.defaultTab)
    if (resolved !== activeTab.value) {
      activeTab.value = resolved
    }
  }

  onMounted(() => window.addEventListener('storage', onStorage))
  onUnmounted(() => window.removeEventListener('storage', onStorage))

  return { activeTab }
}
