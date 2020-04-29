export default {
  template: `
    <span :class="'badge u-badge-' + kind">
      <i :class="'far fa-' + icons[kind]"></i>
      {{ kindText }}
    </span>
  `,

  props: {
    initKind: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      kind: this.initKind.toLowerCase(),
      icons: {
        great: 'heart',
        good: 'thumbs-up',
        average: 'meh',
        bad: 'thumbs-down',
      },
    };
  },

  computed: {
    kindText() {
      return AnnConfig.i18n.ratingState[this.kind];
    },
  },
};
