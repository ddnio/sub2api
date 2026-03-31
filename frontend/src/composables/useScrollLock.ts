/**
 * 全局 scroll-lock 管理器（引用计数）
 *
 * 使用 position: fixed 冻结页面，避免 overflow: hidden 触发全页重绘。
 * 保存/恢复滚动位置，用户无感知。
 */
let lockCount = 0
let savedScrollY = 0

function lock() {
  lockCount++
  if (lockCount === 1) {
    savedScrollY = window.scrollY
    const body = document.body
    body.style.position = 'fixed'
    body.style.top = `-${savedScrollY}px`
    body.style.left = '0'
    body.style.right = '0'
  }
}

function unlock() {
  lockCount = Math.max(0, lockCount - 1)
  if (lockCount === 0) {
    const body = document.body
    body.style.position = ''
    body.style.top = ''
    body.style.left = ''
    body.style.right = ''
    window.scrollTo(0, savedScrollY)
  }
}

/** 强制重置（仅用于异常兜底） */
function forceUnlock() {
  const wasLocked = lockCount > 0
  lockCount = 0
  const body = document.body
  body.style.position = ''
  body.style.top = ''
  body.style.left = ''
  body.style.right = ''
  if (wasLocked) {
    window.scrollTo(0, savedScrollY)
  }
}

export function useScrollLock() {
  return { lock, unlock, forceUnlock }
}
