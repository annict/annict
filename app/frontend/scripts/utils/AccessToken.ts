import axios from 'axios'

export default {
  async generate() {
    const accessToken = localStorage.getItem('accessToken')

    if (accessToken) {
      return
    }

    const res = await axios.post('/api/internal/v3/access_token')

    localStorage.setItem('accessToken', res.data.accessToken)
  },
}
