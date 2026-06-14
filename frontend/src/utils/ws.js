export function resolveWsBaseUrl() {
  const configured = import.meta.env.VITE_WS_URL
  if (configured) {
    return configured.replace(/\/$/, '')
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}`
}
