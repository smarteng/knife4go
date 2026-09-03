<template>
  <div class="document">
    <a-row style="margin-top:10px;">
      <a-button type="primary" :id="'btnCopyOpenApi' + api.id">
        <CopyOutlined /><span> {{$t('open.copy')}} </span>
      </a-button>

      <a-button style="margin-left:10px;" @click="triggerDownloadOpen">
        <DownloadOutlined /> <span> {{$t('open.download')}} </span>
      </a-button>
    </a-row>
    <div style="margin-top:10px;" :id="'knife4jDocumentOpenApiShowEditor'">
      <editor-show v-model:value="openApiRaw" theme="eclipse"></editor-show>
    </div>
  </div>
</template>
<script>
import KUtils from "@/core/utils";
import ClipboardJS from "clipboard";
import { VAceEditor } from 'vue3-ace-editor'
import { computed, defineAsyncComponent } from 'vue'
import { CopyOutlined, DownloadOutlined } from '@ant-design/icons-vue'
import { useGlobalsStore } from '@/store/modules/global.js'
import { useI18n } from 'vue-i18n'
import { message } from 'ant-design-vue'

import ace from "ace-builds";
import modeJson from "ace-builds/src-noconflict/mode-json?url";
import modeJson5 from "ace-builds/src-noconflict/mode-json5.js?url";
import modeXml from "ace-builds/src-noconflict/mode-xml?url";
import themeEclipse from "ace-builds/src-noconflict/theme-eclipse?url";
import extLanguageTools from "ace-builds/src-noconflict/ext-language_tools?url";

ace.config.setModuleUrl('ace/mode/json', modeJson)
ace.config.setModuleUrl('ace/mode/json', modeJson5)
ace.config.setModuleUrl('ace/mode/xml', modeXml)
ace.config.setModuleUrl('ace/theme/eclipse', themeEclipse)
ace.config.setModuleUrl('ace/ext-language/tools', extLanguageTools)

export default {
  name: "Document",
  components: {
    editor: VAceEditor,
    CopyOutlined,
    DownloadOutlined,
    "EditorShow": defineAsyncComponent(() => import('./EditorShow.vue'))
  },
  props: {
    api: {
      type: Object,
      required: true
    },
    swaggerInstance: {
      type: Object,
      required: true
    }
  },
  setup() {
    const store = useGlobalsStore()
    const language = computed(() => {
      return store.language
    })

    const { messages } = useI18n()
    return {
      language,
      messages
    }
  },
  data() {
    return {
      openApiRaw: "",
      name: "OpenAPI.json"
    }
  },
  created() {
    this.openApiRaw = KUtils.json5stringify(this.api.openApiRaw);
    this.name = this.api.summary + "_OpenAPI.json";
    // console.log(this.api);
    // 保存已创建的 ClipboardJS 实例，供组件销毁时释放
    this.$_clipboards = [];
  },
  mounted() {
    // DOM 挂载后再绑定 clipboard，避免 setTimeout 硬等导致的时序不可靠
    this.$nextTick(() => {
      this.copyOpenApi();
    });
  },
  beforeUnmount() {
    // 释放 ClipboardJS 实例，防止内存泄漏
    if (Array.isArray(this.$_clipboards)) {
      this.$_clipboards.forEach(c => c && c.destroy && c.destroy());
      this.$_clipboards = [];
    }
  },
  methods: {
    getCurrentI18nInstance() {
      return this.messages[this.language];
    },
    triggerDownloadOpen() {
      var content = this.openApiRaw;
      var a = document.createElement("a");
      // var content = this.getHtmlContent(this.data.instance.title);
      var option = {};
      var fileName = this.name;
      var url = window.URL.createObjectURL(
        new Blob([content], {
          type:
            (option.type || "text/plain") +
            ";charset=" +
            (option.encoding || "utf-8")
        })
      );
      a.href = url;
      a.download = fileName || "file";
      a.click();
      window.URL.revokeObjectURL(url);
    },
    copyOpenApi() {
      const btnId = "btnCopyOpenApi" + this.api.id
      const el = document.getElementById(btnId)
      if (!el) return
      const clipboard = new ClipboardJS(el, {
        text: () => {
          return this.openApiRaw;
        }
      })
      clipboard.on("success", () => {
        const inst = this.getCurrentI18nInstance()
        // "复制地址成功"
        const successMessage = inst.message.copy.open.success
        message.info(successMessage);
      })
      clipboard.on("error", (e) => {
        console.log(e)
        const inst = this.getCurrentI18nInstance()
        console.log(inst)
        // "复制地址失败"
        const failMessage = inst.message.copy.open.fail
        message.info(failMessage);
      });
      this.$_clipboards.push(clipboard);
    }
  }
}
</script>