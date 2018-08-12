import Ajax from './Ajax'

export default {
  async generate() {
    const accessToken = localStorage.getItem('accessToken')

    if (accessToken) {
      return
    }

    const res = await Ajax.fetch('/api/internal/v3/access_token', { method: 'POST' })

    localStorage.setItem('accessToken', res.data.accessToken)
  },
}
