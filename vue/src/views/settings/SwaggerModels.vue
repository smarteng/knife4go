<template>
  <a-layout-content class="knife4j-body-content">
    <div class="swaggermododel">
      <a-collapse v-model:activeKey="activeKeys" @change="modelChange">
        <a-collapse-panel v-for="model in modelNames" :key="model.id" :class="model.modelClass()">
          <!-- 自定义 header slot: 内嵌一个带 data-model-anchor 属性的 <span> 作为可靠的滚动定位锚点。
               相比于 $refs (依赖 Vue 组件实例) 和 :id (antd-vue 不透传),
               data-* + querySelector 不依赖组件实例, 即使 tab 切换时组件重建也能定位。 -->
          <template #header>
            <span :data-model-anchor="model.id">{{ model.name }}</span>
          </template>
          <a-table v-if="model.load" :columns="columns" :dataSource="model.data"
                   :rowKey="(r) => r.id + r.name" size="middle" :pagination="page">
<!--            <template #expandedRowRender="{ column, record }">-->
<!--              {{ `column` + column}}-->
<!--            </template>-->
          </a-table>
        </a-collapse-panel>
      </a-collapse>
    </div>
  </a-layout-content>
</template>
<script>
import KUtils from "@/core/utils";
import Constants from "@/store/constants";
import { computed, ref, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useGlobalsStore } from '@/store/modules/global.js'
import { useI18n } from 'vue-i18n'
import { useknife4jModels } from '@/store/knife4jModels.js'

export default {
  props: {
    data: {
      type: Object
    }
  },
  setup(props) {
    const expanRows = ref(true)
    const page = ref(false)
    const modelNames = ref([])
    // activeKeys 作为 v-model:activeKey 双向绑定属性供 <a-collapse> 使用,
    // 程序性替换即可展开对应面板, 同时会触发 modelChange 回调完成懒加载。
    // antd-vue 3.x 的 <a-collapse> 使用 v-model:activeKey (单数), 与 1.x 的 v-model 不同。
    const activeKeys = ref([])

    const { messages } = useI18n()

    const globalsStore = useGlobalsStore()
    const swagger = computed(() => {
      return globalsStore.swagger
    })

    const columns = computed(() => {
      return  messages.value[globalsStore.language].table.swaggerModelsColumns;
    })

    const Knife4jModels = useknife4jModels()
    // useRoute() 在 setup() 中获取当前路由对象, 用于:
    //   1) mounted 时首次从 query.model 读取目标并定位
    //   2) watch 监听后续导航变化 (同一组件实例复用时的每次跳转)
    const route = useRoute()
    function initModelNames() {
      const key = Constants.globalTreeTableModelParams + props.data.instance.id
      // 根据instance的实例初始化model名称
      const treeTableModel = props.data.instance.swaggerTreeTableModels
      Knife4jModels.setValue(key, treeTableModel);
      if (KUtils.checkUndefined(treeTableModel)) {
        for (const name in treeTableModel) {
          const random = parseInt(Math.random() * (6 - 1 + 1) + 1, 10)
          const modelInfo = {
            // id: KUtils.randomMd5Str(name),
            id: name,
            name: name,
            // 是否加载过
            load: false,
            data: [],
            random: random
          }
          modelInfo.modelClass = function () {
            let cname = "panel-default"
            switch (random) {
              case 1:
                cname = "panel-success";
                break;
              case 2:
                cname = "panel-success";
                break;
              case 3:
                cname = "panel-info";
                break;
              case 4:
                cname = "panel-warning";
                break;
              case 5:
                cname = "panel-danger";
                break;
              case 6:
                cname = "panel-default";
                break;
            }
            return cname;
          };
          modelNames.value.push(modelInfo);
        }
      }
      // console.log(this.modelNames)
    }

    function modelChange(key) {
      // console("当前激活面板key:" + that.activeKey);

      const instanceKey =
          Constants.globalTreeTableModelParams + props.data.instance.id
      // console("chang事件-------");
      // console(key);

      if (KUtils.arrNotEmpty(key)) {
        // 默认要取最后一个
        const lastIndex = key.length - 1
        const id = key[lastIndex]
        // console("key------------");
        modelNames.value.forEach(function (model) {
          if (model.id == id) {
            // console("找到匹配的model了===");
            // 找到该model,判断是否已加载
            if (!model.load) {
              // 未加载的情况下,进行查找数据
              // //console("查找属性");
              // //console(model);
              const modelData = []
              // 得到当前model的原始对象
              // 所有丶属性全部深拷贝,pid设置为-1
              // var originalModel = treeTableModel[model.name];
              let originalModel = Knife4jModels.getByModelName(
                  instanceKey,
                  model.name
              )
              originalModel = swagger.value.analysisDefinitionRefTableModel(props.data.instance.id, originalModel);
              // console.log("初始化完成")
              console.log(originalModel.children);
              // console("查找原始model:" + model.name);
              if (KUtils.checkUndefined(originalModel)) {
                // 存在
                // 查找属性集合
                if (KUtils.arrNotEmpty(originalModel.params)) {
                  originalModel.params.forEach(function (nmd) {
                    // 第一层属性的pid=-1
                    const childrenParam = {
                      children: nmd.children,
                      childrenTypes: nmd.childrenTypes,
                      def: nmd.def,
                      description: nmd.description,
                      enum: nmd.enum,
                      example: nmd.example,
                      id: nmd.id,
                      ignoreFilterName: nmd.ignoreFilterName,
                      in: nmd.in,
                      level: nmd.level,
                      name: nmd.name,
                      parentTypes: nmd.parentTypes,
                      pid: "-1",
                      readOnly: nmd.readOnly,
                      require: nmd.require,
                      schema: nmd.schema,
                      schemaValue: nmd.schemaValue,
                      show: nmd.show,
                      txtValue: nmd.txtValue,
                      type: nmd.type,
                      validateInstance: nmd.validateInstance,
                      validateStatus: nmd.validateStatus,
                      value: nmd.value
                    }
                    modelData.push(childrenParam);
                    // 判断是否存在schema
                  });
                }
              }
              // //console(modelData);
              model.data = modelData;
              model.load = true;
            }
          }
        });

        console.log(modelNames.value)
      }
      // 第二次复制
      expanRows.value = true;
    }

    initModelNames()
    watch(() => modelNames.value, () => {
      for (let model of modelNames.value) {
        console.log(model.data)
      }
    })

    // focusModel 定位到指定名称的 Model 面板: 展开 + 懒加载 + 精确滚动到锚点。
    // 核心难点: knife4j 采用多 tab 架构, tab 切换时可能销毁重建组件, 导致:
    //   1) $nextTick 回调可能不执行
    //   2) BasicLayout.freePanelMemory 会把非激活 tab 的 props.data.instance 置为 null,
    //      导致 modelChange 内访问 props.data.instance.id 抛 TypeError
    // 综合策略:
    //   a) 先从 store 恢复 props.data.instance, 避免 modelChange 崩溃。
    //   b) 使用 document.querySelector('[data-model-anchor=xxx]') 直接查 DOM, 不依赖组件实例。
    //   c) 使用 window.setTimeout + requestAnimationFrame 递归重试, 不使用 nextTick。
    //   d) 连续两次 rectHeight 相同且 > 60px 才判定高度稳定, 避免展开动画未完成就滚动。
    function focusModel(modelName) {
      if (!modelName) {
        return;
      }
      let target = null;
      for (let i = 0; i < modelNames.value.length; i++) {
        if (modelNames.value[i].id === modelName) {
          target = modelNames.value[i];
          break;
        }
      }
      if (target == null) {
        return;
      }
      // 守卫 props.data.instance: BasicLayout.freePanelMemory 会在切换 tab 时把非激活 tab 的 instance 置为 null。
      // 若通过 router.push 触发 tab 切换, BasicLayout 不会主动恢复 instance,
      // 会导致 modelChange 内访问 null.id 抛 TypeError。这里主动从 store 恢复。
      if (props.data && !props.data.instance) {
        const storeInstance = globalsStore.swaggerCurrentInstance;
        if (storeInstance) {
          props.data.instance = storeInstance;
        }
      }
      // 每次跳转仅展开目标面板一个, 避免多次跳转后面板累积展开挤压页面。
      activeKeys.value = [target.id];
      // 手动触发懒加载, 以兼容 v-model 直接赋值不会同步触发 @change 的情况。
      modelChange(activeKeys.value);
      // 递归重试滚动: 每帧检查目标 DOM 是否已完全渲染, 高度稳定后再精确滚动。
      const maxRetries = 40;
      let retryCount = 0;
      const anchorSelector = '[data-model-anchor="' + target.id + '"]';
      let lastHeight = -1;
      let initialScrollDone = false;
      const tryScroll = function () {
        retryCount++;
        const anchor = document.querySelector(anchorSelector);
        let el = anchor;
        if (anchor && typeof anchor.closest === 'function') {
          const panelRoot = anchor.closest('.ant-collapse-item');
          if (panelRoot) {
            el = panelRoot;
          }
        }
        const rect = el && typeof el.getBoundingClientRect === 'function' ? el.getBoundingClientRect() : null;
        const rectHeight = rect ? rect.height : 0;
        const heightStable = rectHeight > 60 && rectHeight === lastHeight;
        lastHeight = rectHeight;
        // 首次拿到元素时做一次粗略滚动, 让 tab 内容先进入视口。
        if (el && !initialScrollDone) {
          initialScrollDone = true;
          el.scrollIntoView({ behavior: 'auto', block: 'start' });
        }
        // 高度稳定后做最终精确滚动。
        if (el && heightStable) {
          el.scrollIntoView({ behavior: 'smooth', block: 'start' });
          return;
        }
        if (retryCount >= maxRetries) {
          if (el) {
            el.scrollIntoView({ behavior: 'smooth', block: 'start' });
          }
          return;
        }
        // 纯浏览器 API: setTimeout(0) + rAF, 不受 Vue 组件生命周期影响。
        window.setTimeout(function () {
          window.requestAnimationFrame(tryScroll);
        }, 0);
      };
      window.setTimeout(function () {
        window.requestAnimationFrame(tryScroll);
      }, 0);
    }

    // 从当前路由 query.model 读取目标并定位, 仅用于组件首次挂载。
    function focusModelFromRoute() {
      const modelName = route && route.query && route.query.model;
      if (modelName) {
        focusModel(modelName);
      }
    }

    // 监听整个路由对象: knife4j 采用多 tab 架构, SwaggerModels 组件一旦创建会长期存活,
    // 后续多次 router.push 都会复用同一组件实例, 需要用 watch 而非 onMounted 响应。
    watch(() => route.fullPath, () => {
      if (!route.path || route.path.indexOf('/SwaggerModels/') !== 0) {
        return;
      }
      const modelName = route.query && route.query.model;
      if (modelName) {
        focusModel(modelName);
      }
    })

    // 首次挂载时 (刷新、外链直达、首次从文档页点击 schema)
    // watch 不会触发 (没有旧 route 对比), 需在 onMounted 中主动定位一次。
    onMounted(() => {
      focusModelFromRoute();
    })

    return {
      columns,
      expanRows,
      page,
      modelNames,
      activeKeys,
      swagger,
      modelChange
    }
  }
};
</script>
<style lang="less" scoped>
@ColHeaderSize: 16px;
@ColTopHeight: 3px;

.swaggermododel {
  width: 98%;
  margin: 20px auto;
}

.ant-collapse {
  .panel-info {
    font-size: @ColHeaderSize;
    background: #bce8f1;
    margin-top: @ColTopHeight;
  }

  .panel-default {
    font-size: @ColHeaderSize;
    background: #ddd;
    margin-top: @ColTopHeight;
  }

  .panel-danger {
    font-size: @ColHeaderSize;
    background: #ebccd1;
    margin-top: @ColTopHeight;
  }

  .panel-success {
    font-size: @ColHeaderSize;
    background: #d6e9c6;
    margin-top: @ColTopHeight;
  }

  .panel-warning {
    font-size: @ColHeaderSize;
    background: #faebcc;
    margin-top: @ColTopHeight;
  }
}
</style>
