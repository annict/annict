import Vue from 'vue'
import Component from 'vue-class-component'

@Component({
  template: '#t-home',
})
export default class Home extends Vue {
  public isComponentLoaded = false

  get isSignedIn() {
    return this.$root.isSignedIn
  }

  get isLoaded() {
    return this.$root.isAppLoaded && this.isComponentLoaded
  }

  public created() {
    this.isComponentLoaded = true
  }
}
