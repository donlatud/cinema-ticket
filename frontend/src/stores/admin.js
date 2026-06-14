import { defineStore } from 'pinia'
import { fetchAdminBookings, fetchAuditLogs } from '@/api/admin'
import { fetchShowtimes } from '@/api/showtimes'

export const useAdminStore = defineStore('admin', {
  state: () => ({
    movies: [],
    bookings: [],
    auditLogs: [],
    loadingBookings: false,
    loadingAudit: false,
    bookingError: null,
    auditError: null,
    hasQueriedBookings: false,
  }),

  actions: {
    async loadMovies() {
      const showtimes = await fetchShowtimes()
      const seen = new Map()
      for (const showtime of showtimes) {
        if (!seen.has(showtime.movie_id)) {
          seen.set(showtime.movie_id, {
            id: showtime.movie_id,
            title: showtime.movie_title,
          })
        }
      }
      this.movies = [...seen.values()]
    },

    async loadBookings(filters) {
      this.loadingBookings = true
      this.bookingError = null
      try {
        this.bookings = await fetchAdminBookings(filters)
        this.hasQueriedBookings = true
      } catch (err) {
        this.bookingError = err.response?.data?.error || 'Failed to load bookings.'
        this.bookings = []
      } finally {
        this.loadingBookings = false
      }
    },

    async loadAuditLogs(filters) {
      this.loadingAudit = true
      this.auditError = null
      try {
        this.auditLogs = await fetchAuditLogs(filters)
      } catch (err) {
        this.auditError = err.response?.data?.error || 'Failed to load audit logs.'
        this.auditLogs = []
      } finally {
        this.loadingAudit = false
      }
    },
  },
})
