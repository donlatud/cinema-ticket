import api from './client'

export async function fetchShowtimes() {
  const { data } = await api.get('/showtimes')
  return data.showtimes
}

export async function fetchSeats(showtimeId) {
  const { data } = await api.get(`/showtimes/${showtimeId}/seats`)
  return data.seats
}

export async function lockSeats(showtimeId, seatNos) {
  const { data } = await api.post(`/showtimes/${showtimeId}/seats/lock`, { seat_nos: seatNos })
  return data.booking
}
