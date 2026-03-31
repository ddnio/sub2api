/**
 * 全局 scroll-lock 管理器（引用计数）
 *
 * 多个弹窗各自调 lock()/unlock()，只有所有弹窗都关闭后才真正恢复滚动。
 * 统一通过 body.modal-open class 控制，不再直接写 style.overflow。
 */
let lockCount = 0

function lock() {
  lockCount++
  if (lockCount === 1) {
    document.body.classList.add('modal-open')
  }
}

function unlock() {
  lockCount = Math.max(0, lockCount - 1)
  if (lockCount === 0) {
    document.body.classList.remove('modal-open')
  }
}

/** 强制重置（仅用于异常兜底） */
function forceUnlock() {
  lockCount = 0
  document.body.classList.remove('modal-open')
}

export function useScrollLock() {
  return { lock, unlock, forceUnlock }
}
