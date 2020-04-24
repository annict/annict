import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="c-sidebar">
      <div class="c-sidebar__background" v-on:click="hideSidebar" v-show="isSidebarOpen"></div>
      <transition name="slide">
        <div class="c-sidebar__content" v-show="isSidebarOpen">
          <slot></slot>
        </div>
      </transition>
    </div>
  `,

  data() {
    return {
      isSidebarOpen: false,
    };
  },

  methods: {
    hideSidebar() {
      this.isSidebarOpen = false;
    },
  },

  created() {
    eventHub.$on('sidebar:show', () => {
      this.isSidebarOpen = true;
    });
  },
};
