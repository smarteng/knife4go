<template>
  <div>
    <editor :value="value" @init="editorInit" :lang="lang" @input="change" theme="eclipse" width="100%"
      :style="{height: editorHeight + 'px'}"></editor>
  </div>
</template>

<script>
import { VAceEditor } from 'vue3-ace-editor'
import ace from "ace-builds";
// Vite 环境下 vue3-ace-editor 不使用 window.ace 全局实例,
// 必须用 ?url 拿到构建后 module 文件的 URL, 通过 ace.config.setModuleUrl 显式注册,
// 否则 ace 内部 require("ace/mode/xxx") 返回 undefined, 访问 .Mode 时崩溃
// (堆栈: ext-language_tools -> Cannot read properties of undefined (reading 'Mode'))。
// 参考 EditorShow.vue 中的相同注册模式。
import modeJavascript from "ace-builds/src-noconflict/mode-javascript?url";
import modeTypescript from "ace-builds/src-noconflict/mode-typescript?url";
import themeEclipse from "ace-builds/src-noconflict/theme-eclipse?url";
import extLanguageTools from "ace-builds/src-noconflict/ext-language_tools?url";

ace.config.setModuleUrl('ace/mode/javascript', modeJavascript)
ace.config.setModuleUrl('ace/mode/typescript', modeTypescript)
ace.config.setModuleUrl('ace/theme/eclipse', themeEclipse)
ace.config.setModuleUrl('ace/ext/language_tools', extLanguageTools)

export default {
  name: "EditorShow",
  components: { editor: VAceEditor },
  props: {
    value: {
      type: [String, Object],
      required: true,
      default: ""
    },
    tsMode: {
      type: Boolean,
      required: false,
      default: false,
    }
  },
  emits: ['showDescription', 'change'],
  data() {
    return {
      lang: "javascript",
      editor: null,
      editorHeight: 200
    };
  },
  methods: {
    resetEditorHeight() {
      var that = this;
      //  重设高度
      setTimeout(() => {
        var length_editor = that.editor.session.getLength();
        if (length_editor == 1) {
          length_editor = 10;
        }
        var rows_editor = length_editor * 16;
        that.editorHeight = rows_editor;
      }, 300);
    },
    change(value) {
      // this.value = value;
      // 重设高度
      // this.resetEditorHeight();
      this.$emit("change", value);
    },
    editorInit(editor) {
      var that = this;
      this.editor = editor;
      // mode/theme/ext 已在模块顶部通过 ace.config.setModuleUrl 全局注册,
      // 无需再手动 require;组件初始化只需要根据 props 设置语言即可。
      if (this.tsMode) {
        this.lang = "typescript";
      }
      // 重设高度
      this.resetEditorHeight();
      this.editor.renderer.on("afterRender", function () {
        that.$emit("showDescription", "123")
      });
    }
  }
};
</script>
