import Vue from 'vue'
import Component from 'vue-class-component'

import eventHub from '../utils/eventHub'

@Component({
  props: {
    initWithSidebar: {
      default: true,
      type: Boolean,
    },
  },
})
export default class DisplayableSidebar extends Vue {
  private mixinValue = 'DisplayableSidebar'
  private withSidebar = this.initWithSidebar

  private mounted() {
    eventHub.$on('sidebar:toggle', () => {
      this.withSidebar = !this.withSidebar
    })
  }
}
