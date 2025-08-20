<script setup lang="ts">
import type { ISearchOptions } from '@xterm/addon-search'
import type { ShellParams } from '~/pages/application/inference/components/type'
import { CanvasAddon } from '@xterm/addon-canvas'
import { FitAddon } from '@xterm/addon-fit'
import { SearchAddon } from '@xterm/addon-search'
import { WebglAddon } from '@xterm/addon-webgl'
import { Terminal } from '@xterm/xterm'
import { isNil, pickBy } from 'lodash'
import { CaseSensitive, Regex, WholeWord } from 'lucide-vue-next'
import { onWatcherCleanup } from 'vue'
import {
  DEFAULT_COLS,
  DEFAULT_ROWS,
  ERROR_CODE_MESSAGE_MAP,
} from '~/pages/application/inference/components/exec/constant'
import { buildBasicUrl, buildUrlWithParams, getPlatform } from '~/pages/application/inference/components/exec/utils'
import { AttachAddon } from './attachAddon'
import './index.scss'
import '@xterm/xterm/css/xterm.css'

interface IProps {
  params: ShellParams
}

const props = defineProps<IProps>()
const isDev = process.env.NODE_ENV === 'development'

const shellUrl = computed(() => {
  const host = window.location.host
  const { baseUrl, params } = buildBasicUrl(props.params)
  const url = `ws${isDev ? '' : 's'}:${host}${baseUrl}`
  return buildUrlWithParams(url, {
    ...params,
    ...pickBy(params, cv => !isNil(cv)),
  })
})

const isWebgl2Supported = (() => {
  let isSupported = window.WebGL2RenderingContext ? undefined : false
  return () => {
    if (isSupported === undefined) {
      const canvas = document.createElement('canvas')
      const gl = canvas.getContext('webgl2', {
        depth: false,
        antialias: false,
      })
      isSupported = gl instanceof window.WebGL2RenderingContext
    }
    return isSupported
  }
})()

const termContainerRef = useTemplateRef('termContainer')
const fitAddonRef = shallowRef(new FitAddon())
const searchAddonRef = shallowRef(new SearchAddon())
const attachAddonRef = shallowRef<AttachAddon>()

const termRef = shallowRef<Terminal>(
  new Terminal({
    fontFamily: 'Menlo, Monaco, "Courier New", monospace',
    fontWeight: 400,
    fontSize: 14,
    cols: DEFAULT_COLS,
    rows: DEFAULT_ROWS,
    allowProposedApi: true,
    theme: {
      background: 'rgb(28, 32, 36)',
    },
  }),
)

const fitSize = () => {
  const { cols = DEFAULT_COLS, rows = DEFAULT_ROWS }
    = fitAddonRef.value.proposeDimensions() ?? {}
  attachAddonRef.value?.sendSizeData({
    cols,
    rows,
  })
  fitAddonRef.value.fit()
}

const { open, close } = useWebSocket(shellUrl.value, {
  onConnected: (ws) => {
    attachAddonRef.value = new AttachAddon(ws, {
      bidirectional: true,
      keepAlive: true,
    })
    termRef.value?.loadAddon(attachAddonRef.value)
    // Fix a tricky issue that causing prompt not showing up
    setTimeout(() => {
      fitSize()
    }, 400)
    setTimeout(() => {
      fitSize()
    }, 1000)
  },
  onDisconnected: (ws, ev) => {
    termRef.value?.writeln(
      `WebSocket \x1B[1;3;31m${ev?.type}\x1B[0m code:${ev.code}`,
    )
    if (ERROR_CODE_MESSAGE_MAP[ev.code] && termRef.value) {
      ERROR_CODE_MESSAGE_MAP[ev.code]
        ?.split('\n')
        .forEach(l => termRef.value?.writeln(l))
      return
    }
    termRef.value?.writeln(`reason:${ev.reason}`)
  },
  onError: (ws, ev) => {
    termRef.value?.writeln(`WebSocket \x1B[1;3;31m${ev?.type}\x1B[0m`)
  },
  autoReconnect: false,
  immediate: false,
})
const showSearchInput = ref(false)
const searchOptions = ref<ISearchOptions>({
  regex: false,
  wholeWord: false,
  caseSensitive: false,
  decorations: {
    matchBackground: '#515C6A',
    matchOverviewRuler: '#A0A0A0CC',
    activeMatchBackground: '#D09057',
    activeMatchColorOverviewRuler: '#000',
  },
})
const searchResult = ref<{ resultIndex: number, resultCount: number }>()
watchEffect(
  () => {
    const observer = new ResizeObserver(() => {
      fitSize()
    })
    const parentEl = termContainerRef.value?.parentElement
    const term = termRef.value
    term.attachCustomKeyEventHandler((ev) => {
      const key = ev.key
      if (key === 'f' && ev.type === 'keydown') {
        const platform = getPlatform()
        if (platform === 'Mac' && ev.metaKey) {
          ev.stopPropagation()
          ev.preventDefault()
          showSearchInput.value = !showSearchInput.value
        }
        if (platform === 'Win' && ev.ctrlKey) {
          ev.stopPropagation()
          ev.preventDefault()
          showSearchInput.value = !showSearchInput.value
        }
      }
      return true
    })
    const unlisten = searchAddonRef.value.onDidChangeResults(
      res => (searchResult.value = res),
    )
    if (parentEl) {
      term.open(termContainerRef.value)
      term.loadAddon(fitAddonRef.value)
      term.loadAddon(searchAddonRef.value)
      open()
      if (isWebgl2Supported()) {
        const webglAddon = new WebglAddon()
        webglAddon.onContextLoss(() => {
          webglAddon.dispose()
        })
        term.loadAddon(webglAddon)
      }
      else {
        term.loadAddon(new CanvasAddon())
      }
      observer.observe(parentEl)
    }

    onWatcherCleanup(() => {
      observer.disconnect()
      close()
      unlisten.dispose()
      term.dispose()
    })
  },
  {
    flush: 'post',
  },
)

// const uploadVisible = ref(false)
// const downloadVisible = ref(false)
const searchKeyword = ref<string>()
const findNext = (val = '') => {
  searchAddonRef.value.findNext(val, searchOptions.value)
}
const findPrev = (val = '') => {
  searchAddonRef.value.findPrevious(val, searchOptions.value)
}
const handleSearch = useDebounceFn(findNext, 300)
watch(searchOptions, () => {
  searchAddonRef.value.clearDecorations()
  handleSearch(searchKeyword.value)
})
const inputRef = useTemplateRef<HTMLInputElement>('input')
watch(showSearchInput, (value) => {
  if (value) {
    inputRef.value?.focus()
  }
  else {
    searchKeyword.value = ''
    handleSearch('')
  }
})
</script>

<template>
  <div ref="termContainer" class="terminal-container">
    <div class="find" :class="[showSearchInput ? 'open' : '']">
      <div class="flex w-full max-w-sm items-center gap-1.5">
        <Input
          ref="input"
          v-model="searchKeyword"
          placeholder="find"
          class="input"
          autofocus
          @update:model-value="(payload) => handleSearch(payload.toString())"
        />
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Toggle v-model:pressed="searchOptions.caseSensitive" size="sm">
                <CaseSensitive class="h-4 w-4" />
              </Toggle>
            </TooltipTrigger>
            <TooltipContent>
              <p>Case sensitive</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Toggle v-model:pressed="searchOptions.wholeWord" size="sm">
                <WholeWord class="h-4 w-4" />
              </Toggle>
            </TooltipTrigger>
            <TooltipContent>
              <p>Whole word</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger>
              <Toggle v-model:pressed="searchOptions.regex" size="sm">
                <Regex class="h-4 w-4" />
              </Toggle>
            </TooltipTrigger>
            <TooltipContent>
              <p>Regex</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
        <div>
          {{
            searchResult?.resultCount
              ? `${searchResult.resultIndex + 1}/${searchResult.resultCount}`
              : `No Result`
          }}
        </div>
        <Button size="icon" @click="findPrev(searchKeyword)">
          <i class="h-4 w-4 codicon-arrow-up" />
        </Button>
        <Button size="icon" @click="findNext(searchKeyword)">
          <i class="h-4 w-4 codicon-arrow-down" />
        </Button>
      </div>
    </div>
  </div>
</template>
