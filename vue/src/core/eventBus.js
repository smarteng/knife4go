// eventBus 是一个零依赖的极简 EventBus,用于替换 Vue 2 的 $root.$on/$off/$emit。
//
// 背景: Vue 3 移除了实例上的 $on/$off/$emit 事件总线 API (RFC-0020),
// 需要迁移到:
//   a) Pinia store (适合状态型通信,但语义变化)
//   b) mitt 库 (Vue 3 官方推荐,但需增加依赖)
//   c) 自建 EventBus (零依赖,语义与原代码 100% 对齐)
// 本文件采用方案 c: 用 Map<eventName, Set<Handler>> 实现;
// API 命名保持 on/off/emit 三件套,方便从 $root.$on/$off/$emit 直接映射迁移。
//
// 使用示例:
//   import eventBus from '@/core/eventBus'
//   // 订阅
//   eventBus.on('some-event', handler)
//   // 触发
//   eventBus.emit('some-event', payload)
//   // 取消订阅 (组件卸载时必须调用,否则会内存泄漏)
//   eventBus.off('some-event', handler)

// listeners 按事件名分桶存储 handler 集合。
// 使用 Set 而非数组: (1) 天然去重 (2) O(1) 删除。
const listeners = new Map()

// on 订阅指定事件,返回取消订阅函数便于函数式写法。
function on(event, handler) {
  if (typeof event !== 'string' || typeof handler !== 'function') {
    return () => {}
  }
  if (!listeners.has(event)) {
    listeners.set(event, new Set())
  }
  listeners.get(event).add(handler)
  return () => off(event, handler)
}

// off 取消订阅。handler 缺省时清空该事件的所有订阅者。
function off(event, handler) {
  const bucket = listeners.get(event)
  if (!bucket) {
    return
  }
  if (handler) {
    bucket.delete(handler)
  } else {
    bucket.clear()
  }
  if (bucket.size === 0) {
    listeners.delete(event)
  }
}

// emit 触发事件。同步顺序调用所有 handler,任一 handler 抛出异常不影响其他 handler。
function emit(event, payload) {
  const bucket = listeners.get(event)
  if (!bucket || bucket.size === 0) {
    return
  }
  // 复制一份快照,避免 handler 内部调用 on/off 影响正在迭代的集合。
  const snapshot = Array.from(bucket)
  for (const handler of snapshot) {
    try {
      handler(payload)
    } catch (err) {
      // 单个 handler 异常不阻断其他 handler,与 Vue 2 $emit 行为一致。
      console.error('[eventBus] handler error for event', event, err)
    }
  }
}

export default { on, off, emit }
