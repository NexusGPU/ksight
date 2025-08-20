import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import Icons from 'unplugin-icons/vite'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  plugins: [
    vue(),
    // Auto import components and APIs
    AutoImport({
      imports: [
        'vue',
        'vue-router',
        'pinia',
        '@vueuse/core',
        'vue-i18n',
        { 'vue-sonner': ['toast'] },
        { '@/shared/composables/confirm.ts': ['useConfirm'] },
      ],
      dts: 'src/auto-imports.d.ts',
      dirs: ['src/shared/composables', 'src/shared/stores'],
      vueTemplate: true, // Enable auto-import in Vue template
      eslintrc: {
        enabled: true,
      },
    }),
    // Auto register components
    Components({
      dirs: ['src/components', 'src/shared/components', 'src/core/layout'],
      dts: 'src/components.d.ts',
    }),
    // Icons support
    Icons({
      autoInstall: true,
    }),
    // i18n support
    VueI18nPlugin({
      include: path.resolve(__dirname, './src/locales/**'),
    }),
  ],
  // For tailwindcss JIT mode
  optimizeDeps: {
    include: ['vue', 'pinia', 'vue-router', 'vue-i18n'],
  },
})
