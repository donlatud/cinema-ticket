<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAdminStore } from '@/stores/admin'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()
const admin = useAdminStore()

const activeTab = ref('bookings')

const bookingFilters = reactive({
  movieId: '',
  date: '',
  userId: '',
})

const auditFilters = reactive({
  event: '',
  from: '',
  to: '',
})

const auditEvents = [
  { value: '', label: 'All Events' },
  { value: 'BOOKING_SUCCESS', label: 'Booking Success' },
  { value: 'BOOKING_TIMEOUT', label: 'Booking Timeout' },
  { value: 'SEAT_RELEASED', label: 'Seat Released' },
  { value: 'SYSTEM_ERROR', label: 'System Error' },
]

const hasBookingFilter = computed(
  () => Boolean(bookingFilters.movieId || bookingFilters.date || bookingFilters.userId),
)

onMounted(() => {
  admin.loadMovies()
})

function applyBookingFilters() {
  if (!hasBookingFilter.value) {
    admin.bookingError = 'Select at least one filter (movie, date, or user) to search.'
    return
  }
  admin.loadBookings({ ...bookingFilters })
}

function clearBookingFilters() {
  bookingFilters.movieId = ''
  bookingFilters.date = ''
  bookingFilters.userId = ''
  admin.bookings = []
  admin.bookingError = null
  admin.hasQueriedBookings = false
}

function loadAudit() {
  admin.loadAuditLogs({ ...auditFilters })
}

function selectTab(tab) {
  activeTab.value = tab
  if (tab === 'audit' && admin.auditLogs.length === 0) {
    loadAudit()
  }
}

function statusClass(status) {
  return {
    PAID: 'badge badge--paid',
    PENDING: 'badge badge--pending',
    EXPIRED: 'badge badge--expired',
  }[status] || 'badge'
}

function shortRef(id) {
  return id ? `BK-${id.slice(-6).toUpperCase()}` : '—'
}

function formatDateTime(value) {
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

function signOut() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="admin">
    <aside class="admin__sidebar">
      <div class="admin__brand">
        <span class="admin__brand-icon" aria-hidden="true">🎬</span>
        <div>
          <p class="admin__brand-name">Cinema Admin</p>
          <p class="admin__brand-sub">Operational View</p>
        </div>
      </div>

      <nav class="admin__nav" aria-label="Admin sections">
        <router-link :to="{ name: 'home' }" class="admin__nav-link">Showtimes</router-link>
        <router-link :to="{ name: 'my-bookings' }" class="admin__nav-link">My Bookings</router-link>
        <span class="admin__nav-link admin__nav-link--active">Admin</span>
      </nav>

      <button type="button" class="admin__signout" @click="signOut">Sign Out</button>
    </aside>

    <main class="admin__main">
      <header class="admin__header">
        <h1 class="admin__title">System Administration</h1>
        <p class="admin__subtitle">Manage bookings, monitor system health, and audit operational logs.</p>
      </header>

      <div class="admin__tabs" role="tablist">
        <button
          type="button"
          role="tab"
          :aria-selected="activeTab === 'bookings'"
          class="admin__tab"
          :class="{ 'admin__tab--active': activeTab === 'bookings' }"
          @click="selectTab('bookings')"
        >
          Bookings Management
        </button>
        <button
          type="button"
          role="tab"
          :aria-selected="activeTab === 'audit'"
          class="admin__tab"
          :class="{ 'admin__tab--active': activeTab === 'audit' }"
          @click="selectTab('audit')"
        >
          Audit Logs
        </button>
      </div>

      <section v-if="activeTab === 'bookings'" class="admin__panel">
        <form class="admin__filters" @submit.prevent="applyBookingFilters">
          <label class="admin__field">
            <span class="admin__field-label">Movie Title</span>
            <select v-model="bookingFilters.movieId" class="admin__input">
              <option value="">All Movies</option>
              <option v-for="movie in admin.movies" :key="movie.id" :value="movie.id">
                {{ movie.title }}
              </option>
            </select>
          </label>

          <label class="admin__field">
            <span class="admin__field-label">Date</span>
            <input v-model="bookingFilters.date" type="date" class="admin__input" />
          </label>

          <label class="admin__field">
            <span class="admin__field-label">User ID</span>
            <input
              v-model="bookingFilters.userId"
              type="text"
              class="admin__input"
              placeholder="e.g. 65f1c2..."
            />
          </label>

          <div class="admin__filter-actions">
            <button type="submit" class="admin__apply">Apply Filters</button>
            <button type="button" class="admin__clear" @click="clearBookingFilters">Clear</button>
          </div>
        </form>

        <p v-if="admin.loadingBookings" class="admin__state" aria-live="polite">Loading bookings...</p>

        <p v-else-if="admin.bookingError" class="admin__state admin__state--error" aria-live="assertive">
          {{ admin.bookingError }}
        </p>

        <p v-else-if="!admin.hasQueriedBookings" class="admin__state">
          Select a filter and apply to view bookings.
        </p>

        <p v-else-if="admin.bookings.length === 0" class="admin__state">
          No bookings match the selected filters.
        </p>

        <div v-else class="admin__table-wrap">
          <table class="admin__table">
            <thead>
              <tr>
                <th scope="col">Booking Ref</th>
                <th scope="col">Movie</th>
                <th scope="col">Showtime</th>
                <th scope="col">Seats</th>
                <th scope="col">Status</th>
                <th scope="col">Amount</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="booking in admin.bookings" :key="booking.id">
                <td class="admin__ref">{{ shortRef(booking.id) }}</td>
                <td>{{ booking.movie_title }}</td>
                <td>{{ formatDateTime(booking.start_time) }}</td>
                <td>{{ booking.seat_nos.join(', ') }}</td>
                <td><span :class="statusClass(booking.status)">{{ booking.status }}</span></td>
                <td>${{ booking.amount.toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section v-else class="admin__panel">
        <form class="admin__filters" @submit.prevent="loadAudit">
          <label class="admin__field">
            <span class="admin__field-label">Event Type</span>
            <select v-model="auditFilters.event" class="admin__input">
              <option v-for="option in auditEvents" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </label>

          <label class="admin__field">
            <span class="admin__field-label">From Date</span>
            <input v-model="auditFilters.from" type="date" class="admin__input" />
          </label>

          <label class="admin__field">
            <span class="admin__field-label">To Date</span>
            <input v-model="auditFilters.to" type="date" class="admin__input" />
          </label>

          <div class="admin__filter-actions">
            <button type="submit" class="admin__apply">Apply Filters</button>
          </div>
        </form>

        <p v-if="admin.loadingAudit" class="admin__state" aria-live="polite">Loading audit logs...</p>

        <p v-else-if="admin.auditError" class="admin__state admin__state--error" aria-live="assertive">
          {{ admin.auditError }}
        </p>

        <p v-else-if="admin.auditLogs.length === 0" class="admin__state">No audit logs found.</p>

        <div v-else class="admin__table-wrap">
          <table class="admin__table">
            <thead>
              <tr>
                <th scope="col">Event Type</th>
                <th scope="col">Seat</th>
                <th scope="col">Detail</th>
                <th scope="col">Created At</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="log in admin.auditLogs" :key="log.id">
                <td class="admin__event">{{ log.event }}</td>
                <td>{{ log.seat_no || '—' }}</td>
                <td>{{ log.detail || '—' }}</td>
                <td>{{ formatDateTime(log.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </main>
  </div>
</template>
