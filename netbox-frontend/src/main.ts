import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { vCan } from './directives/can'
import { useAuthStore } from './stores/auth'
import './style.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.directive('can', vCan)

window.addEventListener('netbox:unauthorized', () => {
  const auth = useAuthStore(pinia)
  auth.invalidateSession()
  if (router.currentRoute.value.name !== 'login') {
    void router.push({ name: 'login', query: { next: router.currentRoute.value.fullPath } })
  }
})

app.mount('#app')
