import $ from 'jquery';

import eventHub from '../../common/eventHub';

const NO_SELECT = 'no_select';

export default {
  template: '#t-status-selector',

  data() {
    return {
      appData: {},
      pageData: {},
      statuses: [],
      isLoading: false,
      statusKind: null,
      prevStatusKind: null,
    };
  },

  props: {
    workId: {
      type: Number,
      required: true,
    },

    size: {
      type: String,
      default: 'default',
    },

    isTransparent: {
      type: Boolean,
      default: false,
    },

    initStatusKind: {
      type: String,
    },
  },

  methods: {
    currentStatusKind() {
      if (!this.statuses.length) {
        return NO_SELECT;
      }

      const status = this.statuses.filter(status => {
        return status.work_id === this.workId;
      })[0];

      if (!status) {
        return NO_SELECT;
      }

      return status.kind;
    },

    resetKind() {
      return (this.statusKind = NO_SELECT);
    },

    change() {
      if (!this.$root.appData.isUserSignedIn) {
        $('.c-sign-up-modal').modal('show');
        this.resetKind();
        return;
      }

      if (this.statusKind !== this.prevStatusKind) {
        this.isLoading = true;

        $.ajax({
          method: 'POST',
          url: `/api/internal/works/${this.workId}/statuses/select`,
          data: {
            status_kind: this.statusKind,
            page_category: gon.page.category,
          },
        }).done(() => {
          this.isLoading = false;
        });
      }
    },
  },

  mounted() {
    this.isLoading = true;

    if (this.initStatusKind) {
      this.prevStatusKind = this.initStatusKind;
      this.statusKind = this.initStatusKind;
      this.isLoading = false;
      return;
    }

    eventHub.$on('app:loaded', () => {
      if (!this.$root.appData.isUserSignedIn) {
        this.statusKind = this.prevStatusKind = NO_SELECT;
        this.isLoading = false;
        return;
      }

      this.statuses = this.$root.pageData.statuses || [];

      if (this.statuses.length) {
        this.prevStatusKind = this.currentStatusKind();
        this.statusKind = this.currentStatusKind();
        this.isLoading = false;
      }
    });
  },
};
