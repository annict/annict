import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="c-tab-bar">
      <slot
        :showSidebar="showSidebar"
      ></slot>
    </div>
  `,

  methods: {
    showSidebar() {
      eventHub.$emit('sidebar:show');
    },
  },
};
