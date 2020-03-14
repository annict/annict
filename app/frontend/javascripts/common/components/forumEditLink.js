import eventHub from '../../common/eventHub';

export default {
  template: '#t-forum-edit-link',

  props: {
    userId: {
      type: Number,
      required: true,
    },
    path: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      appData: {},
      appLoaded: false,
    };
  },

  methods: {
    isEditable: function() {
      return this.userId === (this.appData.current_user && this.appData.current_user.id);
    },
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$root.appData;
      this.appLoaded = true;
    });
  },
};
