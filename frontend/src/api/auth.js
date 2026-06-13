import api from './client'

export async function login(idToken) {
  const { data } = await api.post('/api/auth/login', {
    id_token: idToken,
  })
  return data
}
