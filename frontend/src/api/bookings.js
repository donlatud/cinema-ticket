import api from './client'

export async function fetchMyBookings() {
  const { data } = await api.get('/api/bookings/my')
  return data.bookings
}

export async function payBooking(bookingId) {
  const { data } = await api.post(`/api/bookings/${bookingId}/pay`)
  return data.booking
}

export async function cancelBooking(bookingId) {
  const { data } = await api.post(`/api/bookings/${bookingId}/cancel`)
  return data.booking
}
