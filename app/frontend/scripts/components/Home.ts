export default {
  template: '#t-home',

  data() {
    return {
      isComponentLoaded: false,
    }
  },

  computed: {
    isSignedIn() {
      return this.$root.isSignedIn
    },

    isLoaded() {
      return this.$root.isAppLoaded && this.isComponentLoaded
    },
  },

  created() {
    this.isComponentLoaded = true
  },
}
