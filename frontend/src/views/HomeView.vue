<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { fetchShowtimes } from '@/api/showtimes'

const router = useRouter()

const showtimes = ref([])
const loading = ref(true)
const error = ref(null)
const search = ref('')

onMounted(async () => {
  try {
    showtimes.value = await fetchShowtimes()
  } catch {
    error.value = 'Failed to load showtimes.'
  } finally {
    loading.value = false
  }
})

const movies = computed(() => {
  const grouped = new Map()

  for (const showtime of showtimes.value) {
    if (!grouped.has(showtime.movie_id)) {
      grouped.set(showtime.movie_id, {
        id: showtime.movie_id,
        title: showtime.movie_title,
        poster: showtime.movie_poster,
        durationMin: showtime.duration_min,
        screen: showtime.screen,
        minPrice: showtime.price,
        showtimes: [],
      })
    }

    const entry = grouped.get(showtime.movie_id)
    entry.showtimes.push(showtime)
    entry.minPrice = Math.min(entry.minPrice, showtime.price)
  }

  for (const entry of grouped.values()) {
    entry.showtimes.sort((a, b) => new Date(a.start_time) - new Date(b.start_time))
  }

  return [...grouped.values()]
})

const filteredMovies = computed(() => {
  const term = search.value.trim().toLowerCase()
  if (!term) {
    return movies.value
  }
  return movies.value.filter((movie) => movie.title.toLowerCase().includes(term))
})

function formatShowtime(value) {
  return new Date(value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function goToSeatMap(showtimeId) {
  router.push({ name: 'seatmap', params: { id: showtimeId } })
}
</script>

<template>
  <section class="showtimes">
    <header class="showtimes__header">
      <div>
        <h2 class="showtimes__title">Today's Showtimes</h2>
        <p class="showtimes__subtitle">Select a movie and time to begin booking.</p>
      </div>
      <label class="showtimes__search">
        <span class="sr-only">Search movies</span>
        <input
          v-model="search"
          type="search"
          class="showtimes__search-input"
          placeholder="Search movies..."
        />
      </label>
    </header>

    <p v-if="loading" class="showtimes__state" aria-live="polite">Loading showtimes...</p>

    <p v-else-if="error" class="showtimes__state showtimes__state--error" aria-live="assertive">
      {{ error }}
    </p>

    <p v-else-if="filteredMovies.length === 0" class="showtimes__state" aria-live="polite">
      No showtimes match your search.
    </p>

    <ul v-else class="showtimes__list">
      <li v-for="movie in filteredMovies" :key="movie.id">
        <article class="movie-card">
          <img
            v-if="movie.poster"
            :src="movie.poster"
            :alt="`${movie.title} poster`"
            class="movie-card__poster"
          />
          <div v-else class="movie-card__poster movie-card__poster--placeholder" aria-hidden="true">
            🎞️
          </div>

          <div class="movie-card__body">
            <h3 class="movie-card__title">{{ movie.title }}</h3>
            <p class="movie-card__meta">
              <span v-if="movie.durationMin">⏱ {{ movie.durationMin }} min</span>
              <span>🎬 {{ movie.screen }}</span>
            </p>

            <ul class="movie-card__times">
              <li v-for="showtime in movie.showtimes" :key="showtime.id">
                <button type="button" class="movie-card__time" @click="goToSeatMap(showtime.id)">
                  <span class="movie-card__time-label">{{ formatShowtime(showtime.start_time) }}</span>
                  <span class="movie-card__time-price">${{ showtime.price.toFixed(2) }}</span>
                </button>
              </li>
            </ul>
          </div>

          <div class="movie-card__action">
            <p class="movie-card__price-label">Starting from</p>
            <p class="movie-card__price">${{ movie.minPrice.toFixed(2) }}</p>
            <button
              type="button"
              class="movie-card__cta"
              @click="goToSeatMap(movie.showtimes[0].id)"
            >
              Select Seats
            </button>
          </div>
        </article>
      </li>
    </ul>
  </section>
</template>
