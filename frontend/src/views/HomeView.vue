<script setup>
import { onMounted, ref } from 'vue'
import { fetchShowtimes } from '@/api/showtimes'

const showtimes = ref([])
const loading = ref(true)
const error = ref(null)

onMounted(async () => {
  try {
    showtimes.value = await fetchShowtimes()
  } catch {
    error.value = 'Failed to load showtimes'
  } finally {
    loading.value = false
  }
})

function formatTime(value) {
  return new Date(value).toLocaleString()
}
</script>

<template>
  <section>
    <header>
      <h2 class="text-2xl font-semibold">Showtimes</h2>
      <p class="mt-2 text-gray-600">Pick a showtime to view the live seat map.</p>
    </header>

    <section v-if="loading" class="mt-6" aria-live="polite">
      <p class="text-gray-600">Loading showtimes...</p>
    </section>

    <section v-else-if="error" class="mt-6" aria-live="assertive">
      <p class="text-red-600">{{ error }}</p>
    </section>

    <section v-else class="mt-6 space-y-4">
      <article
        v-for="showtime in showtimes"
        :key="showtime.id"
        class="rounded-lg border border-gray-200 bg-white p-4 shadow-sm"
      >
        <h3 class="text-lg font-medium">{{ showtime.movie_title }}</h3>
        <p class="mt-1 text-gray-600">{{ showtime.screen }}</p>
        <p class="text-gray-600">{{ formatTime(showtime.start_time) }}</p>
        <p class="mt-1 font-medium">${{ showtime.price.toFixed(2) }}</p>
        <router-link
          :to="{ name: 'seatmap', params: { id: showtime.id } }"
          class="mt-3 inline-block text-indigo-600 hover:text-indigo-800"
        >
          View seats
        </router-link>
      </article>
    </section>
  </section>
</template>
