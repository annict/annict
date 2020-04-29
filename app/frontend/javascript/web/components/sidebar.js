import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="c-sidebar">
      <slot
        v-bind:hideSidebar="hideSidebar"
        v-bind:isSidebarOpen="isSidebarOpen"
      ></slot>
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
