export default {
  async setup() {
    const res = await fetch('/api/internal/v3/base_data')
    const data = await res.json()

    const baseData = {
      csrfParam: data.csrf.param,
      csrfToken: data.csrf.token,
      domain: data.domain,
      env: data.env,
      locale: data.locale,
    }

    for (const key in baseData) {
      localStorage.setItem(key, baseData[key])
    }
  },

  async fetch(keyName) {
    const val = localStorage.getItem(keyName)

    if (val) {
      return val
    }

    await this.setup()

    return localStorage.getItem(keyName)
  },
}
