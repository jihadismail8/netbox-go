import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const sidebarCollapsed = ref(localStorage.getItem('netbox_sidebar') === 'collapsed')
  const darkMode = ref(localStorage.getItem('netbox_theme') === 'dark')

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
    localStorage.setItem('netbox_sidebar', sidebarCollapsed.value ? 'collapsed' : 'expanded')
  }

  function toggleDarkMode() {
    darkMode.value = !darkMode.value
    localStorage.setItem('netbox_theme', darkMode.value ? 'dark' : 'light')
    document.documentElement.classList.toggle('dark', darkMode.value)
  }

  function initTheme() {
    document.documentElement.classList.toggle('dark', darkMode.value)
  }

  return { sidebarCollapsed, darkMode, toggleSidebar, toggleDarkMode, initTheme }
})
