export default {
  setup(options) {
    this.baseOptions = options
  },

  async fetch(url, options = {}) {
    const res = await fetch(url, Object.assign(options, this.baseOptions))

    return {
      data: await res.json(),
    }
  },
}
