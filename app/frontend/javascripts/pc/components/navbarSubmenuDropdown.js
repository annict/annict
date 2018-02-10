import eventHub from '../../common/eventHub'

export default {
  template: '#t-navbar-submenu-dropdown',

  data() {
    return {
      appData: {},
      appLoaded: false,
    }
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$parent.appData
      this.appLoaded = true
    })
  },
}
