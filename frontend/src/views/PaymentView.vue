<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useBookingStore } from '@/stores/booking'

const route = useRoute()
const router = useRouter()
const bookingStore = useBookingStore()

const bookingId = computed(() => route.params.id)
const now = ref(Date.now())
let timer = null

const booking = computed(() => bookingStore.current)

const showtime = computed(() => {
  if (!booking.value) {
    return null
  }
  return bookingStore.showtimeMap[booking.value.showtime_id] || null
})

const isPending = computed(() => booking.value?.status === 'PENDING')

const remainingMs = computed(() => {
  if (!booking.value?.expires_at) {
    return 0
  }
  return Math.max(0, new Date(booking.value.expires_at).getTime() - now.value)
})

const isExpired = computed(() => isPending.value && remainingMs.value <= 0)

const countdown = computed(() => {
  const totalSeconds = Math.floor(remainingMs.value / 1000)
  const minutes = String(Math.floor(totalSeconds / 60)).padStart(2, '0')
  const seconds = String(totalSeconds % 60).padStart(2, '0')
  return `${minutes}:${seconds}`
})

const startTimeLabel = computed(() => {
  if (!showtime.value?.start_time) {
    return '—'
  }
  return new Date(showtime.value.start_time).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
})

onMounted(async () => {
  await bookingStore.loadBooking(bookingId.value)
  timer = setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  if (timer) {
    clearInterval(timer)
  }
})

async function handlePay() {
  const result = await bookingStore.pay(bookingId.value)
  if (result) {
    router.push({ name: 'my-bookings' })
  }
}

async function handleCancel() {
  const result = await bookingStore.cancel(bookingId.value)
  if (result) {
    router.push({ name: 'home' })
  }
}
</script>

<template>
  <div class="page-dark">
    <article class="payment-card">
      <header class="payment-card__header">
        <h1 class="payment-card__title">Payment</h1>
        <p class="payment-card__subtitle">Complete your booking for</p>
        <p class="payment-card__movie">{{ showtime?.movie_title || 'Your booking' }}</p>
      </header>

      <section v-if="bookingStore.loading" class="payment-card__state" aria-live="polite">
        <p>Loading booking...</p>
      </section>

      <section v-else-if="!booking" class="payment-card__state" aria-live="assertive">
        <p>Booking not found.</p>
        <button type="button" class="payment-card__secondary" @click="router.push({ name: 'home' })">
          Back to Showtimes
        </button>
      </section>

      <template v-else>
        <dl class="payment-summary">
          <div class="payment-summary__row">
            <dt>Showtime</dt>
            <dd>{{ startTimeLabel }}</dd>
          </div>
          <div class="payment-summary__row">
            <dt>Screen</dt>
            <dd>{{ showtime?.screen || '—' }}</dd>
          </div>
          <div class="payment-summary__row">
            <dt>Seats</dt>
            <dd>{{ booking.seat_nos.join(', ') }}</dd>
          </div>
          <div class="payment-summary__row payment-summary__row--total">
            <dt>Total</dt>
            <dd>${{ booking.amount.toFixed(2) }}</dd>
          </div>
        </dl>

        <section v-if="isPending && !isExpired" class="payment-timer" aria-live="polite">
          <p class="payment-timer__label">Time remaining to complete booking</p>
          <p class="payment-timer__value">⏱ {{ countdown }}</p>
        </section>

        <p
          v-if="isExpired"
          class="payment-alert payment-alert--error"
          aria-live="assertive"
        >
          This booking has expired. Please select your seats again.
        </p>

        <p v-if="booking.status === 'PAID'" class="payment-alert payment-alert--success">
          Payment complete. Enjoy the movie!
        </p>

        <p v-if="bookingStore.error" class="payment-alert payment-alert--error" aria-live="assertive">
          {{ bookingStore.error }}
        </p>

        <div v-if="isPending && !isExpired" class="payment-actions">
          <button
            type="button"
            class="payment-actions__pay"
            :disabled="bookingStore.processing"
            @click="handlePay"
          >
            {{ bookingStore.processing ? 'Processing...' : 'Pay (Mock)' }}
          </button>
          <button
            type="button"
            class="payment-actions__cancel"
            :disabled="bookingStore.processing"
            @click="handleCancel"
          >
            Cancel Booking
          </button>
        </div>

        <button
          v-else
          type="button"
          class="payment-card__secondary"
          @click="router.push({ name: 'my-bookings' })"
        >
          Go to My Bookings
        </button>
      </template>
    </article>
  </div>
</template>
