import Vue from 'vue'
import Component from 'vue-class-component'

import eventHub from '../utils/eventHub'

@Component({
  template: '#t-navbar',
})
export default class Navbar extends Vue {
  private toggleSidebar() {
    eventHub.$emit('sidebar:toggle')
  }
}
