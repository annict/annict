import eventHub from '../../common/eventHub'

export default {
  template: '#t-navbar-left-menu',

  props: {
    isTransparent: {
      type: Boolean,
      default: false,
    },
  },

  data() {
    return {
      appData: {},
      appLoaded: false,
    }
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$root.appData
      this.appLoaded = true
    })
  },
}
