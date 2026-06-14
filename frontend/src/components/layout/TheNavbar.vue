<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

function handleSignOut() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <header class="app-navbar">
    <div class="app-navbar__inner">
      <nav class="app-navbar__nav" aria-label="Primary">
        <router-link :to="{ name: 'home' }" class="app-navbar__brand">
          <span class="app-navbar__brand-icon" aria-hidden="true">🎬</span>
          <span>Cinema Booking</span>
        </router-link>

        <ul class="app-navbar__links">
          <li>
            <router-link :to="{ name: 'home' }" class="app-navbar__link">Showtimes</router-link>
          </li>
          <li>
            <router-link :to="{ name: 'my-bookings' }" class="app-navbar__link">My Bookings</router-link>
          </li>
          <li v-if="auth.isAdmin">
            <router-link :to="{ name: 'admin' }" class="app-navbar__link">Admin</router-link>
          </li>
        </ul>
      </nav>

      <div class="app-navbar__actions">
        <button type="button" class="app-navbar__signout" @click="handleSignOut">Sign Out</button>
        <span class="app-navbar__avatar" :title="auth.displayName" aria-hidden="true">
          {{ auth.initials }}
        </span>
      </div>
    </div>
  </header>
</template>
