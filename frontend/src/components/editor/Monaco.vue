<!-- eslint-disable no-unused-vars -->
<script lang="ts" setup>
import * as monaco from 'monaco-editor'
import { configureMonacoYaml } from 'monaco-yaml'

const props = withDefaults(defineProps<Props>(), {
  lang: () => 'yaml',
  options: () => ({}),
  modelValue: () => '',
})

const emit = defineEmits<Emits>()

const mode = useColorMode()

interface Props {
  /**
   * Programming Language (Not a locale for UI);
   * overrides `options.language`
   */
  lang?: string
  /**
   * Options passed to the second argument of `monaco.editor.create`
   */
  options?: monaco.editor.IStandaloneEditorConstructionOptions
  /**
   * The URI that identifies models
   */
  modelUri?: monaco.Uri

  modelValue?: string

  schemaUrl?: string
}

interface Emits {
  (event: 'update:modelValue', value: string): void
  (event: 'load', editor: monaco.editor.IStandaloneCodeEditor): void
  (event: 'onInput', value: string): void
}

const lang = computed(() => props.lang || props.options.language)
const editorRef = shallowRef<monaco.editor.IStandaloneCodeEditor>()
const editorElement = ref<HTMLDivElement>()
const defaultOptions: monaco.editor.IStandaloneEditorConstructionOptions = {
  automaticLayout: true,
  theme: mode.value === 'dark' ? 'vs-dark' : 'vs-light',
}

let editor: monaco.editor.IStandaloneCodeEditor
let model: monaco.editor.ITextModel

watch(() => props.modelValue, () => {
  if (editor?.getValue() !== props.modelValue) {
    editor?.setValue(props.modelValue)
  }
})

watch(() => [props.lang, props.modelUri], () => {
  if (model) {
    model.dispose()
  }
  model = monaco.editor.createModel(props.modelValue, lang.value, props.modelUri)
  editor?.setModel(model)
})

watch(() => [props.options, mode], () => {
  editor?.updateOptions({ ...defaultOptions, theme: mode.value === 'dark' ? 'vs-dark' : 'vs-light', ...props.options })
})

watch(editorElement, () => {
  if (!editorElement.value)
    return
  props.schemaUrl && configureMonacoYaml(monaco, {
    validate: true,
    enableSchemaRequest: true,
    format: true,
    hover: true,
    completion: true,
    schemas: [
      {
        fileMatch: ['*'],
        uri: props.schemaUrl,
      },
    ],
  })
  editor = monaco.editor.create(editorElement.value, { ...defaultOptions, ...props.options })
  model = monaco.editor.createModel(props.modelValue, lang.value, props.modelUri)
  editorRef.value = editor
  editor.layout()
  editor.setModel(model)
  editor.onDidChangeModelContent(() => {
    emit('update:modelValue', editor.getValue())
    emit('onInput', editor.getValue())

    // Trigger Suggest
    // const position = editor.getPosition()
    // const model = editor.getModel()
    // if (model && position) {
    //   const word = model.getWordAtPosition(position)
    //   if (word && /^[a-z]+$/i.test(word.word)) {
    //     editor.trigger('keyboard', 'editor.action.triggerSuggest', {})
    //   }
    // }
  })
  emit('load', editor)

  monaco.languages.registerCompletionItemProvider('yaml', {
    provideCompletionItems(model, position) {
      const word = model.getWordUntilPosition(position)
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }
      return {
        suggestions: [
          {
            label: '---',
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: '---',
            range,
          },
        ],
      }
    },
    triggerCharacters: ['-'],
  })
})

defineExpose({
  /**
   * Monaco editor instance
   */
  $editor: editorRef,
})

onBeforeUnmount(() => {
  editor?.dispose()
  model?.dispose()
})
</script>

<template>
  <div ref="editorElement" />
</template>
