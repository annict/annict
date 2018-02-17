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
    eventHub.$on('app:loaded', ({ appData }) => {
      this.appData = appData
      this.appLoaded = true
    })
  },
}
