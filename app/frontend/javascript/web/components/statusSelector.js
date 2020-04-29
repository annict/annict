import $ from 'jquery';

import eventHub from '../../common/eventHub';

const NO_SELECT = 'no_select';

export default {
  template: `
    <div
      class="c-status-selector"
      :class='{
        "unselected": statusKind === "no_select",
        "c-spinner": isLoading
      }'>
      <div class="c-status-selector__object">
        <select class="w-100" v-model="statusKind" @change="change">
          <slot></slot>
        </select>
        <i class="fas fa-caret-down"></i>
      </div>
    </div>
  `,

  data() {
    return {
      isLoading: false,
      statusKind: null,
      prevStatusKind: null,
      isUserSignedIn: false,
    };
  },

  props: {
    workId: {
      type: Number,
      required: true,
    },

    initStatusKind: {
      type: String,
    },

    pageCategory: {
      type: String,
    },
  },

  methods: {
    currentStatusKind(libraryEntries) {
      if (!libraryEntries.length) {
        return NO_SELECT;
      }

      const status = libraryEntries.filter((status) => {
        return status.work_id === this.workId;
      })[0];

      if (!status) {
        return NO_SELECT;
      }

      return status.status_kind;
    },

    resetKind() {
      this.statusKind = NO_SELECT;
    },

    change() {
      if (!this.isUserSignedIn) {
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
            page_category: this.pageCategory,
          },
        }).done(() => {
          this.isLoading = false;
        });
      }
    },
  },

  mounted() {
    this.isLoading = true;

    eventHub.$on('request:libraryEntries:fetched', (libraryEntries) => {
      if (!libraryEntries) {
        this.isUserSignedIn = false;
        return;
      }

      this.isUserSignedIn = true;

      if (!libraryEntries || !libraryEntries.length) {
        this.statusKind = this.prevStatusKind = NO_SELECT;
        this.isLoading = false;
        return;
      }

      this.statusKind = this.prevStatusKind = this.currentStatusKind(libraryEntries);
      this.isLoading = false;
    });
  },
};
