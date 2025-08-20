import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import router from '../core/router'
import App from '../App.vue'
import '../style.css'
import en from '../locales/en.json'
import zh from '../locales/zh.json'

// Initialize i18n with configuration
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages: {
    en,
    zh
  }
})

// Create and configure Vue app
const app = createApp(App)

// Install plugins
app.use(createPinia())
app.use(router)
app.use(i18n)

// Mount the app
app.mount('#app')
