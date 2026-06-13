<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const isAuthLayout = computed(() => route.meta.layout === 'auth')
</script>

<template>
  <router-view v-if="isAuthLayout" />

  <div v-else class="min-h-screen bg-gray-50 text-gray-900">
    <header class="border-b border-gray-200 bg-white px-4 py-4 sm:px-6">
      <h1 class="text-lg font-semibold sm:text-xl">Cinema Booking</h1>
    </header>
    <main class="mx-auto max-w-4xl px-4 py-6 sm:px-6 sm:py-8">
      <router-view />
    </main>
  </div>
</template>
