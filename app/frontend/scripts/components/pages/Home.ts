import Vue from 'vue'
import Component from 'vue-class-component'

import GuestHome from '../GuestHome'

@Component({
  components: {
    'c-guest-home': GuestHome,
  },
  template: '#t-home',
})
export default class Home extends Vue {
  private isComponentLoaded = false
  private root: any = this.$root

  get isSignedIn() {
    return this.root.isSignedIn
  }

  get isLoaded() {
    return this.root.isAppLoaded && this.isComponentLoaded
  }

  private created() {
    this.isComponentLoaded = true
  }
}
