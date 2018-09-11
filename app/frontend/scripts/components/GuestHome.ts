import Component, { mixins } from 'vue-class-component'

import DisplayableSidebar from '../mixins/DisplayableSidebar'

import Navbar from './Navbar'
import Sidebar from './Sidebar'

@Component({
  components: {
    'c-navbar': Navbar,
    'c-sidebar': Sidebar,
  },
  template: '#t-guest-home',
})
export default class GuestHome extends mixins(DisplayableSidebar) {}
