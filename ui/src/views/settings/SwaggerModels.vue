<template>
  <a-layout-content class="knife4j-body-content">
    <div class="swaggermododel">
      <a-collapse v-model="activeKeys" @change="modelChange">
        <a-collapse-panel v-for="model in modelNames" :key="model.id" :class="model.modelClass()"
          :ref="'panel_' + model.id">
          <!-- 自定义 header slot: 内嵌一个带 data-model-anchor 属性的 <span>,
               作为可靠的滚动定位锚点。
               相比于 $refs (依赖 Vue 组件实例) 和 :id (antd-vue 1.7.8 不透传),
               data-* + querySelector 不依赖组件实例, 即使 tab 切换时组件重建也能定位。 -->
          <span slot="header" :data-model-anchor="model.id">{{ model.name }}</span>
          <a-table v-if="model.load" :defaultExpandAllRows="expanRows" :columns="columns" :dataSource="model.data"
            :rowKey="unionKey" size="middle" :pagination="page">
            <template slot="descriptionValueTemplate" slot-scope="text">
              <span v-html="text"></span>
            </template>
          </a-table>
        </a-collapse-panel>
      </a-collapse>
    </div>
  </a-layout-content>
</template>
<script>
import KUtils from "@/core/utils";
import Constants from "@/store/constants";

export default {
  props: {
    data: {
      type: Object
    }
  },
  computed: {
    language() {
      return this.$store.state.globals.language;
    },
    swagger() {
      return this.$store.state.globals.swagger;
    }
  },
  watch: {
    language: function (val, oldval) {
      this.initI18n();
    },
    // 监听整个 $route 对象而非 $route.query.model 深路径。
    // 原因: Vue 2 + vue-router 3.1.3 每次导航会将整个 $route 引用整体替换,
    // 监听整个 $route 一定能触发; 而监听 $route.query.model 深路径在某些情况下
    // (如 query 对象新旧引用同时变化) 可能不会可靠触发。
    // knife4j 采用多 tab 架构, SwaggerModels 组件一旦创建会长期存活,
    // 后续多次 $router.push 都会复用同一组件实例, 需要用 watch 而非 mounted 响应。
    $route: function (to) {
      if (!to || !to.path) {
        return;
      }
      // 只处理指向本 SwaggerModels 页的导航, 避免其它路由变化时误触发。
      if (to.path.indexOf('/SwaggerModels/') !== 0) {
        return;
      }
      var modelName = to.query && to.query.model;
      if (modelName) {
        this.focusModel(modelName);
      }
    }
  },
  created() {
    this.initI18n();
    this.initModelNames();
  },
  mounted() {
    // 首次挂载时 (刷新、外链直达、首次从文档页点击 schema)
    // watch 不会触发 (因为没有旧 $route 可供对比), 需在 mounted 中主动定位一次。
    this.focusModelFromRoute();
  },
  data() {
    return {
      columns: [],
      expanRows: true,
      page: false,
      modelNames: [],
      // activeKeys 作为双向绑定属性供 <a-collapse> 使用,
      // 程序性替换即可展开对应面板, 同时会触发 modelChange 回调完成懒加载。
      activeKeys: []
    };
  },
  methods: {
    getCurrentI18nInstance() {
      return this.$i18n.messages[this.language];
    },
    initI18n() {
      this.columns = this.getCurrentI18nInstance().table.swaggerModelsColumns;
    },
    unionKey() {
      return KUtils.randomMd5();
    },
    // focusModelFromRoute 从当前路由 query 中读取 model 名称, 仅用于组件首次挂载。
    focusModelFromRoute() {
      var modelName = this.$route && this.$route.query && this.$route.query.model;
      if (modelName) {
        this.focusModel(modelName);
      }
    },
    // focusModel 定位到指定名称的 Model 面板: 展开 + 懒加载 + 精确滚动到锚点。
    // 核心难点: knife4j 采用多 tab 架构 (<component :is="pane.content">),
    //   1) tab 切换时 antd <a-tabs> 会销毁重建当前 SwaggerModels 组件, Vue 取消 $nextTick 队列,
    //      导致依赖 $nextTick 的回调根本不会执行。
    //   2) 旧组件实例销毁后, that.$refs 失效, 闭包持有的也是旧引用。
    //   3) BasicLayout.freePanelMemory 会把非激活 tab 的 this.data.instance 置为 null,
    //      导致 modelChange 内访问 this.data.instance.id 抛 TypeError, 中止 focusModel。
    // 综合策略:
    //   a) 先从 store 恢复 this.data.instance, 避免 modelChange 崩溃。
    //   b) 使用 document.querySelector('[data-model-anchor=xxx]') 直接查 DOM, 不依赖组件实例。
    //   c) 使用 window.setTimeout + requestAnimationFrame 递归重试, 不使用 this.$nextTick。
    //   d) 连续两次 rectHeight 相同且 > 60px 才判定高度稳定, 避免展开动画未完成就滚动到中间位置。
    focusModel(modelName) {
      if (!modelName) {
        return;
      }
      var target = null;
      for (var i = 0; i < this.modelNames.length; i++) {
        if (this.modelNames[i].id === modelName) {
          target = this.modelNames[i];
          break;
        }
      }
      if (target == null) {
        return;
      }
      // 守卫 this.data.instance: BasicLayout.freePanelMemory 会在切换 tab 时把非激活 tab 的 instance 置为 null。
      // 若通过 $router.push 触发 tab 切换 (走 $route watch 路径), BasicLayout 只调 watchPathMenuSelect
      // 更新菜单, 不调 freePanelMemory 恢复 instance, 会导致 modelChange 内访问 null.id 抛 TypeError。
      // 这里主动从 store 恢复 instance, 保证 modelChange 内部能正确访问。
      if (this.data && !this.data.instance) {
        var storeInstance = this.$store && this.$store.state && this.$store.state.globals && this.$store.state.globals.swaggerCurrentInstance;
        if (storeInstance) {
          this.data.instance = storeInstance;
        }
      }
      // 每次跳转仅展开目标面板一个, 避免多次跳转后面板累积展开挤压页面。
      // 直接替换整个数组以确保 <a-collapse> 的 v-model 触发响应。
      this.activeKeys = [target.id];
      // 手动触发懒加载, 以兼容 v-model 直接赋值不会同步触发 @change 的情况。
      this.modelChange(this.activeKeys);
      // 递归重试滚动: 每帧检查目标 DOM 是否已完全渲染, 高度稳定后再精确滚动。
      var maxRetries = 40;
      var retryCount = 0;
      var anchorSelector = '[data-model-anchor="' + target.id + '"]';
      // 高度稳定判据: 连续两次 rectHeight 相同, 且已明显超过折叠状态的 header 高度 (>60px)。
      var lastHeight = -1;
      var initialScrollDone = false;
      var tryScroll = function () {
        retryCount++;
        // 定位到带 data-model-anchor 的 header span, 然后向上找到 a-collapse-item 根节点作为滚动目标。
        var anchor = document.querySelector(anchorSelector);
        var el = anchor;
        if (anchor && typeof anchor.closest === 'function') {
          var panelRoot = anchor.closest('.ant-collapse-item');
          if (panelRoot) {
            el = panelRoot;
          }
        }
        var rect = el && typeof el.getBoundingClientRect === 'function' ? el.getBoundingClientRect() : null;
        var rectHeight = rect ? rect.height : 0;
        var heightStable = rectHeight > 60 && rectHeight === lastHeight;
        lastHeight = rectHeight;
        // 首次拿到元素时做一次粗略滚动, 让 tab 内容先进入视口 (处理从其它 tab 跳回 SwaggerModels 时视口偏移的情况)。
        if (el && !initialScrollDone) {
          initialScrollDone = true;
          el.scrollIntoView({ behavior: 'auto', block: 'start' });
        }
        // 高度稳定后再做最终精确滚动 (使用 smooth 平滑滚动到最终位置)。
        if (el && heightStable) {
          el.scrollIntoView({ behavior: 'smooth', block: 'start' });
          return;
        }
        if (retryCount >= maxRetries) {
          // 超过重试上限, 若拿到过元素则用当前高度做兜底滚动。
          if (el) {
            el.scrollIntoView({ behavior: 'smooth', block: 'start' });
          }
          return;
        }
        // 不用 this.$nextTick: 它依赖组件实例, 一旦组件被 tab 切换销毁就不执行了。
        // 改用纯浏览器 API: setTimeout(0) + rAF 保证一帧一帧推进, 且不受 Vue 组件生命周期影响。
        window.setTimeout(function () {
          window.requestAnimationFrame(tryScroll);
        }, 0);
      };
      // 首次尝试也不用 nextTick, 直接给浏览器一帧时间完成 patch 后再拿 DOM。
      window.setTimeout(function () {
        window.requestAnimationFrame(tryScroll);
      }, 0);
    },
    initModelNames() {
      var key = Constants.globalTreeTableModelParams + this.data.instance.id;
      // 根据instance的实例初始化model名称
      var treeTableModel = this.data.instance.swaggerTreeTableModels;
      this.$Knife4jModels.setValue(key, treeTableModel);
      if (KUtils.checkUndefined(treeTableModel)) {
        for (var name in treeTableModel) {
          var random = parseInt(Math.random() * (6 - 1 + 1) + 1, 10);
          var modelInfo = {
            // id: KUtils.randomMd5Str(name),
            id: name,
            name: name,
            // 是否加载过
            load: false,
            data: [],
            random: random
          };
          modelInfo.modelClass = function () {
            var cname = "panel-default";
            switch (this.random) {
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
          this.modelNames.push(modelInfo);
        }
      }
      // console.log(this.modelNames)
    },
    modelChange(key) {
      var that = this;
      // console("当前激活面板key:" + that.activeKey);

      var instanceKey =
        Constants.globalTreeTableModelParams + this.data.instance.id;
      // console("chang事件-------");
      // console(key);

      if (KUtils.arrNotEmpty(key)) {
        // 默认要取最后一个
        var lastIndex = key.length - 1;
        var id = key[lastIndex];
        // console("key------------");
        this.modelNames.forEach(function (model) {
          if (model.id == id) {
            // console("找到匹配的model了===");
            // 找到该model,判断是否已加载
            if (!model.load) {
              // 未加载的情况下,进行查找数据
              // //console("查找属性");
              // //console(model);
              var modelData = [];
              // 得到当前model的原始对象
              // 所有丶属性全部深拷贝,pid设置为-1
              // var originalModel = treeTableModel[model.name];
              var originalModel = that.$Knife4jModels.getByModelName(
                instanceKey,
                model.name
              );
              originalModel = that.swagger.analysisDefinitionRefTableModel(that.data.instance.id, originalModel);
              // console.log("初始化完成")
              // console.log(originalModel);
              // console("查找原始model:" + model.name);
              if (KUtils.checkUndefined(originalModel)) {
                // 存在
                // 查找属性集合
                if (KUtils.arrNotEmpty(originalModel.params)) {
                  originalModel.params.forEach(function (nmd) {
                    // 第一层属性的pid=-1
                    var childrenParam = {
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
                    };
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
      }
      // 第二次复制
      that.expanRows = true;
    },
    deepFindChildren(modelData) {
      var that = this;
      var paramDatas = [];
      if (KUtils.arrNotEmpty(modelData)) {
        // 找出第一基本的父级结构
        modelData.forEach(function (md) {
          var newmd = {
            childrenTypes: md.childrenTypes,
            def: md.def,
            description: md.description,
            enum: md.enum,
            example: md.example,
            id: md.id,
            ignoreFilterName: md.ignoreFilterName,
            in: md.in,
            level: md.level,
            name: md.name,
            parentTypes: md.parentTypes,
            pid: md.pid,
            readOnly: md.readOnly,
            require: md.require,
            schema: md.schema,
            schemaValue: md.schemaValue,
            show: md.show,
            txtValue: md.txtValue,
            type: md.type,
            validateInstance: md.validateInstance,
            validateStatus: md.validateStatus,
            value: md.value
          };
          if (newmd.pid == "-1") {
            newmd.children = [];
            newmd.childrenIds = [];
            that.findModelChildren(newmd, modelData);
            // 查找后如果没有,则将children置空
            if (newmd.children.length == 0) {
              newmd.children = null;
            }
            //  modelA.data.push(md)
            paramDatas.push(newmd);
          }
        });
      }
      return paramDatas;
    },
    findModelChildren(md, modelData) {
      var that = this;
      if (KUtils.arrNotEmpty(modelData)) {
        modelData.forEach(function (nmd) {
          var newnmd = {
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
            pid: nmd.pid,
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
          };
          if (newnmd.pid == md.id) {
            newnmd.children = [];
            newnmd.childrenIds = [];
            that.findModelChildren(newnmd, modelData);
            // 查找后如果没有,则将children置空
            if (newnmd.children.length == 0) {
              newnmd.children = null;
            }
            // 判断是否存在
            if (md.childrenIds.indexOf(newnmd.id) == -1) {
              // 不存在
              md.childrenIds.push(newnmd.id);
              md.children.push(newnmd);
            }
          }
        });
      }
    },
    deepTreeTableSchemaModel(modelData, treeTableModel, param, rootParam) {
      var that = this;
      // //console(model.name)
      if (KUtils.checkUndefined(param.schemaValue)) {
        var schema = treeTableModel[param.schemaValue];
        if (KUtils.checkUndefined(schema)) {
          rootParam.parentTypes.push(param.schemaValue);
          if (KUtils.arrNotEmpty(schema.params)) {
            schema.params.forEach(function (nmd) {
              // childrenparam需要深拷贝一个对象
              var childrenParam = {
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
                pid: nmd.pid,
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
              };
              childrenParam.pid = param.id;
              modelData.push(childrenParam);
              if (childrenParam.schema) {
                // 存在schema,判断是否出现过
                if (
                  rootParam.parentTypes.indexOf(childrenParam.schemaValue) == -1
                ) {
                  that.deepTreeTableSchemaModel(
                    modelData,
                    treeTableModel,
                    childrenParam,
                    rootParam
                  );
                }
              }
            });
          }
        }
      }
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
