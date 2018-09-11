import Vue from 'vue'
import Component from 'vue-class-component'

@Component({
  props: {
    show: {
      required: true,
      type: Boolean,
    },
  },
  template: '#t-sidebar',
})
export default class Sidebar extends Vue {}
