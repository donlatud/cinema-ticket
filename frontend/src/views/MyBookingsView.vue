<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useBookingStore } from '@/stores/booking'

const router = useRouter()
const bookingStore = useBookingStore()

const search = ref('')
const statusFilter = ref('ALL')

onMounted(() => {
  bookingStore.loadMyBookings()
})

const filteredBookings = computed(() => {
  const term = search.value.trim().toLowerCase()

  return bookingStore.enrichedBookings.filter((booking) => {
    const matchesStatus = statusFilter.value === 'ALL' || booking.status === statusFilter.value
    const matchesSearch = !term || booking.movie_title.toLowerCase().includes(term)
    return matchesStatus && matchesSearch
  })
})

function shortId(id) {
  return id ? `#${id.slice(-6).toUpperCase()}` : '—'
}

function formatShowtime(value) {
  if (!value) {
    return '—'
  }
  return new Date(value).toLocaleString([], {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function statusClass(status) {
  return {
    PENDING: 'badge badge--pending',
    PAID: 'badge badge--paid',
    EXPIRED: 'badge badge--expired',
  }[status] || 'badge'
}

function completePayment(bookingId) {
  router.push({ name: 'payment', params: { id: bookingId } })
}
</script>

<template>
  <section class="bookings">
    <header class="bookings__header">
      <div>
        <h2 class="bookings__title">My Bookings</h2>
        <p class="bookings__subtitle">Manage your recent and upcoming cinema reservations.</p>
      </div>
      <div class="bookings__filters">
        <label class="sr-only" for="booking-search">Search bookings</label>
        <input
          id="booking-search"
          v-model="search"
          type="search"
          class="bookings__search"
          placeholder="Search bookings..."
        />
        <label class="sr-only" for="status-filter">Filter by status</label>
        <select id="status-filter" v-model="statusFilter" class="bookings__select">
          <option value="ALL">All Statuses</option>
          <option value="PENDING">Pending</option>
          <option value="PAID">Paid</option>
          <option value="EXPIRED">Expired</option>
        </select>
      </div>
    </header>

    <p v-if="bookingStore.loading" class="bookings__state" aria-live="polite">Loading bookings...</p>

    <p v-else-if="bookingStore.error" class="bookings__state bookings__state--error" aria-live="assertive">
      {{ bookingStore.error }}
    </p>

    <p v-else-if="filteredBookings.length === 0" class="bookings__state" aria-live="polite">
      You have no bookings yet.
    </p>

    <div v-else class="bookings__table-wrap">
      <table class="bookings__table">
        <thead>
          <tr>
            <th scope="col">Booking ID</th>
            <th scope="col">Movie</th>
            <th scope="col">Showtime</th>
            <th scope="col">Seats</th>
            <th scope="col">Amount</th>
            <th scope="col">Status</th>
            <th scope="col" class="bookings__col-action">Action</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in filteredBookings" :key="booking.id">
            <td class="bookings__id">{{ shortId(booking.id) }}</td>
            <td>
              <p class="bookings__movie">{{ booking.movie_title }}</p>
              <p class="bookings__screen">{{ booking.screen }}</p>
            </td>
            <td>{{ formatShowtime(booking.start_time) }}</td>
            <td>{{ booking.seat_nos.join(', ') }}</td>
            <td>${{ booking.amount.toFixed(2) }}</td>
            <td><span :class="statusClass(booking.status)">{{ booking.status }}</span></td>
            <td class="bookings__col-action">
              <button
                v-if="booking.status === 'PENDING'"
                type="button"
                class="bookings__action"
                @click="completePayment(booking.id)"
              >
                Complete Payment
              </button>
              <span v-else-if="booking.status === 'PAID'" class="bookings__action-muted">Paid</span>
              <span v-else class="bookings__action-muted">Details</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
