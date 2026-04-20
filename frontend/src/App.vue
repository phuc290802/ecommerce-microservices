<template>
  <div id="app-root">
    <Navbar v-if="auth.isLoggedIn" />
    <main :class="{ 'with-navbar': auth.isLoggedIn }">
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" />
        </transition>
      </router-view>
    </main>
  </div>
</template>

<script setup>
import Navbar from './components/Navbar.vue'
import { useAuthStore } from './stores/auth'
const auth = useAuthStore()
</script>

<style>
#app-root {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

main.with-navbar {
  margin-top: 64px;
}
</style>
