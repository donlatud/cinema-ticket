import { defineStore } from 'pinia'
import { cancelBooking, fetchMyBookings, payBooking } from '@/api/bookings'
import { fetchShowtimes } from '@/api/showtimes'

export const useBookingStore = defineStore('booking', {
  state: () => ({
    current: null,
    bookings: [],
    showtimeMap: {},
    loading: false,
    processing: false,
    error: null,
  }),

  getters: {
    enrichedBookings: (state) => {
      return state.bookings.map((booking) => {
        const showtime = state.showtimeMap[booking.showtime_id] || null
        return {
          ...booking,
          movie_title: showtime?.movie_title || 'Unknown movie',
          screen: showtime?.screen || '—',
          start_time: showtime?.start_time || null,
        }
      })
    },
  },

  actions: {
    setCurrent(booking) {
      this.current = booking
    },

    async loadShowtimeMap() {
      const showtimes = await fetchShowtimes()
      const map = {}
      for (const showtime of showtimes) {
        map[showtime.id] = showtime
      }
      this.showtimeMap = map
    },

    async loadMyBookings() {
      this.loading = true
      this.error = null
      try {
        const [bookings] = await Promise.all([fetchMyBookings(), this.loadShowtimeMap()])
        this.bookings = bookings
      } catch {
        this.error = 'Failed to load your bookings.'
      } finally {
        this.loading = false
      }
    },

    async loadBooking(bookingId) {
      this.loading = true
      this.error = null
      try {
        if (this.current?.id === bookingId && Object.keys(this.showtimeMap).length) {
          return this.current
        }
        const [bookings] = await Promise.all([fetchMyBookings(), this.loadShowtimeMap()])
        this.bookings = bookings
        this.current = bookings.find((item) => item.id === bookingId) || null
        return this.current
      } catch {
        this.error = 'Failed to load booking.'
        return null
      } finally {
        this.loading = false
      }
    },

    async pay(bookingId) {
      this.processing = true
      this.error = null
      try {
        const booking = await payBooking(bookingId)
        this.current = booking
        return booking
      } catch (err) {
        this.error = err.response?.data?.error || 'Payment failed. Please try again.'
        return null
      } finally {
        this.processing = false
      }
    },

    async cancel(bookingId) {
      this.processing = true
      this.error = null
      try {
        const booking = await cancelBooking(bookingId)
        this.current = booking
        return booking
      } catch (err) {
        this.error = err.response?.data?.error || 'Failed to cancel booking.'
        return null
      } finally {
        this.processing = false
      }
    },
  },
})
