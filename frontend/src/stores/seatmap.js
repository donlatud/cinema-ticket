import { defineStore } from 'pinia'
import { fetchSeats, fetchShowtimes, lockSeats } from '@/api/showtimes'
import { resolveWsBaseUrl } from '@/utils/ws'

export const useSeatmapStore = defineStore('seatmap', {
  state: () => ({
    showtime: null,
    seats: [],
    selectedSeatNos: [],
    loading: false,
    locking: false,
    error: null,
    conflictMessage: null,
    ws: null,
  }),

  getters: {
    seatByNo: (state) => {
      const map = {}
      for (const seat of state.seats) {
        map[seat.seat_no] = seat
      }
      return map
    },

    totalPrice: (state) => {
      if (!state.showtime) {
        return 0
      }
      return state.selectedSeatNos.length * state.showtime.price
    },

    selectedSeatsLabel: (state) => {
      return [...state.selectedSeatNos]
        .sort((a, b) => {
          const rowA = a.charAt(0)
          const rowB = b.charAt(0)
          if (rowA !== rowB) {
            return rowA.localeCompare(rowB)
          }
          return parseInt(a.slice(1), 10) - parseInt(b.slice(1), 10)
        })
        .join(', ')
    },
  },

  actions: {
    async loadShowtime(showtimeId) {
      const showtimes = await fetchShowtimes()
      this.showtime = showtimes.find((item) => item.id === showtimeId) ?? null
    },

    async loadSeats(showtimeId) {
      this.loading = true
      this.error = null
      try {
        await this.loadShowtime(showtimeId)
        this.seats = await fetchSeats(showtimeId)
        this.selectedSeatNos = []
        this.conflictMessage = null
      } catch {
        this.error = 'Failed to load seats'
      } finally {
        this.loading = false
      }
    },

    connectWebSocket(showtimeId) {
      this.disconnectWebSocket()

      const baseUrl = resolveWsBaseUrl()
      const ws = new WebSocket(`${baseUrl}/ws?showtime_id=${showtimeId}`)

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          this.handleUpdate(message)
        } catch {
          // ignore malformed messages
        }
      }

      this.ws = ws
    },

    disconnectWebSocket() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
    },

    handleUpdate(message) {
      if (message.type !== 'SEAT_UPDATE') {
        return
      }

      const seat = this.seats.find((item) => item.seat_no === message.seat_no)
      if (seat) {
        seat.status = message.status
      }

      if (message.status !== 'AVAILABLE') {
        this.selectedSeatNos = this.selectedSeatNos.filter((no) => no !== message.seat_no)
      }
    },

    toggleSeat(seatNo) {
      this.conflictMessage = null

      const seat = this.seats.find((item) => item.seat_no === seatNo)
      if (!seat || seat.status !== 'AVAILABLE') {
        return
      }

      if (this.selectedSeatNos.includes(seatNo)) {
        this.selectedSeatNos = this.selectedSeatNos.filter((no) => no !== seatNo)
        return
      }

      this.selectedSeatNos = [...this.selectedSeatNos, seatNo]
    },

    async lockSelected(showtimeId) {
      if (this.selectedSeatNos.length === 0) {
        return null
      }

      this.locking = true
      this.error = null
      this.conflictMessage = null

      try {
        const booking = await lockSeats(showtimeId, this.selectedSeatNos)
        this.selectedSeatNos = []
        return booking
      } catch (err) {
        const status = err.response?.status
        const serverError = err.response?.data?.error || ''

        if (status === 409) {
          const seatMatch = serverError.match(/:\s*([A-J]\d+)/i)
          const seatNo = seatMatch?.[1]?.toUpperCase()
          this.conflictMessage = seatNo
            ? `Seat ${seatNo} was just locked by another terminal. Please select a different seat.`
            : 'One or more seats were just taken. Please select different seats.'
          this.selectedSeatNos = []
          this.seats = await fetchSeats(showtimeId)
        } else if (status === 401) {
          this.error = 'Please sign in to book seats.'
        } else {
          this.error = 'Failed to lock seats. Please try again.'
        }
        return null
      } finally {
        this.locking = false
      }
    },
  },
})
