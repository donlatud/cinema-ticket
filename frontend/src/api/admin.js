import api from './client'

export async function fetchAdminBookings(filters = {}) {
  const params = {}
  if (filters.movieId) params.movie_id = filters.movieId
  if (filters.userId) params.user_id = filters.userId
  if (filters.date) params.date = filters.date

  const { data } = await api.get('/api/admin/bookings', { params })
  return data.bookings
}

export async function fetchAuditLogs(filters = {}) {
  const params = {}
  if (filters.event) params.event = filters.event
  if (filters.from) params.from = filters.from
  if (filters.to) params.to = filters.to

  const { data } = await api.get('/api/admin/audit-logs', { params })
  return data.audit_logs
}
