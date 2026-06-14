<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import SeatGrid from '@/components/SeatGrid.vue'
import SeatMapAlert from '@/components/seatmap/SeatMapAlert.vue'
import SeatMapFooter from '@/components/seatmap/SeatMapFooter.vue'
import SeatMapHeader from '@/components/seatmap/SeatMapHeader.vue'
import SeatMapLegend from '@/components/seatmap/SeatMapLegend.vue'
import SeatMapPageLayout from '@/components/seatmap/SeatMapPageLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useBookingStore } from '@/stores/booking'
import { useSeatmapStore } from '@/stores/seatmap'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const seatmap = useSeatmapStore()
const bookingStore = useBookingStore()

const showtimeId = computed(() => route.params.id)
const successMessage = ref('')

const showtimeMeta = computed(() => {
  const showtime = seatmap.showtime
  if (!showtime) {
    return {
      movieTitle: 'Showtime',
      startTime: '',
      screen: '',
      pricePerSeat: 0,
    }
  }

  return {
    movieTitle: showtime.movie_title || 'Showtime',
    startTime: new Date(showtime.start_time).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
    }),
    screen: showtime.screen,
    pricePerSeat: showtime.price,
  }
})

onMounted(async () => {
  await seatmap.loadSeats(showtimeId.value)
  seatmap.connectWebSocket(showtimeId.value)
})

onUnmounted(() => {
  seatmap.disconnectWebSocket()
})

async function handleLock() {
  if (!auth.token) {
    router.push({ name: 'login', query: { redirect: route.fullPath } })
    return
  }

  successMessage.value = ''
  const booking = await seatmap.lockSelected(showtimeId.value)
  if (booking) {
    bookingStore.setCurrent(booking)
    router.push({ name: 'payment', params: { id: booking.id } })
  }
}
</script>

<template>
  <SeatMapPageLayout>
    <article class="seatmap-card">
      <SeatMapHeader
        :movie-title="showtimeMeta.movieTitle"
        :start-time="showtimeMeta.startTime"
        :screen="showtimeMeta.screen"
        :price-per-seat="showtimeMeta.pricePerSeat"
      />

      <section v-if="seatmap.loading" class="seatmap-card__state" aria-live="polite">
        <p>Loading seat map...</p>
      </section>

      <section v-else-if="seatmap.error" class="seatmap-card__state seatmap-card__state--error" aria-live="assertive">
        <p>{{ seatmap.error }}</p>
      </section>

      <template v-else>
        <SeatMapAlert v-if="seatmap.conflictMessage" :message="seatmap.conflictMessage" />

        <p v-if="successMessage" class="seatmap-success" aria-live="polite">
          {{ successMessage }}
        </p>

        <SeatMapLegend />

        <SeatGrid
          :seats="seatmap.seats"
          :selected-seat-nos="seatmap.selectedSeatNos"
          @toggle="seatmap.toggleSeat"
        />

        <SeatMapFooter
          :selected-count="seatmap.selectedSeatNos.length"
          :selected-label="seatmap.selectedSeatsLabel"
          :total-price="seatmap.totalPrice"
          :locking="seatmap.locking"
          :disabled="seatmap.selectedSeatNos.length === 0"
          @confirm="handleLock"
        />
      </template>
    </article>
  </SeatMapPageLayout>
</template>
