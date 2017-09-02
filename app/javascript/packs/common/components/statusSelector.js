import $ from 'jquery';
import _ from 'lodash';

import eventHub from '../eventHub';

const NO_SELECT = 'no_select';

export default {
  template: '#t-status-selector',

  data() {
    return {
      isLoading: false,
      isSignedIn: gon.user.isSignedIn,
      statusKind: null,
      prevStatusKind: null,
      works: [],
      pageObject: gon.pageObject ? JSON.parse(gon.pageObject) : {}
    };
  },

  props: {
    workId: {
      type: Number,
      required: true
    },

    size: {
      type: String,
      default: 'default'
    },

    isTransparent: {
      type: Boolean,
      default: false
    },

    initStatusKind: {
      type: String
    }
  },

  methods: {
    currentStatusKind() {
      if (!this.works.length) {
        return 'no_select';
      }
      const data = _.find(this.works, work => {
        return work.id === this.workId;
      });
      return data.statusSelector.currentStatusKind;
    },

    resetKind() {
      return (this.statusKind = NO_SELECT);
    },

    change() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        this.resetKind();
        return;
      }

      if (this.statusKind !== this.prevStatusKind) {
        this.isLoading = true;

        return $.ajax({
          method: 'POST',
          url: `/api/internal/works/${this.workId}/statuses/select`,
          data: {
            status_kind: this.statusKind,
            page_category: gon.basic.pageCategory
          }
        }).done(() => {
          return (this.isLoading = false);
        });
      }
    }
  },

  mounted() {
    if (!this.isSignedIn) {
      this.statusKind = this.prevStatusKind = NO_SELECT;
      return;
    }

    if (this.initStatusKind) {
      this.prevStatusKind = this.initStatusKind;
      return (this.statusKind = this.initStatusKind);
    } else {
      this.works = this.pageObject.works;
      this.prevStatusKind = this.currentStatusKind();
      return (this.statusKind = this.currentStatusKind());
    }
  }
};
