<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import TheNavbar from '@/components/layout/TheNavbar.vue'

const route = useRoute()
const layout = computed(() => route.meta.layout ?? 'default')
</script>

<template>
  <router-view v-if="layout === 'auth'" />

  <router-view v-else-if="layout === 'seatmap' || layout === 'plain'" />

  <div v-else-if="layout === 'admin'" class="min-h-screen bg-slate-100">
    <router-view />
  </div>

  <div v-else class="min-h-screen bg-gray-50 text-gray-900">
    <TheNavbar />
    <main class="mx-auto max-w-5xl px-4 py-6 sm:px-6 sm:py-8">
      <router-view />
    </main>
  </div>
</template>
