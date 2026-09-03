// 轻量事件总线：替代 Vue 2 中 this.$root.$on / $off / $emit 的用法
// Vue 3 已移除根实例事件总线 API，这里提供一个最小实现供跨组件通信使用

const listeners = new Map()

/**
 * 订阅事件
 * @param {string} event 事件名
 * @param {Function} handler 事件回调
 */
export function on(event, handler) {
  if (!listeners.has(event)) {
    listeners.set(event, new Set())
  }
  listeners.get(event).add(handler)
}

/**
 * 取消订阅
 * @param {string} event 事件名
 * @param {Function} handler 事件回调（不传则清空该事件所有监听）
 */
export function off(event, handler) {
  const handlers = listeners.get(event)
  if (!handlers) return
  if (handler) {
    handlers.delete(handler)
  } else {
    handlers.clear()
  }
}

/**
 * 发布事件
 * @param {string} event 事件名
 * @param  {...any} args 事件参数
 */
export function emit(event, ...args) {
  const handlers = listeners.get(event)
  if (!handlers) return
  handlers.forEach(handler => {
    try {
      handler(...args)
    } catch (err) {
      console.error(`[eventBus] handler for "${event}" threw:`, err)
    }
  })
}

export default { on, off, emit }
