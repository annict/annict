import $ from 'jquery';

export default {
  template: '#t-reaction-button',

  props: {
    resourceType: {
      type: String,
      required: true,
    },
    resourceId: {
      type: Number,
      required: true,
    },
    initReactionsCount: {
      type: Number,
      required: true,
    },
    initIsReacted: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      isSignedIn: gon.user.isSignedIn,
      reactionsCount: this.initReactionsCount,
      isReacted: this.initIsReacted,
      isLoading: false,
    };
  },

  methods: {
    toggleReact() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }

      if (this.isLoading) {
        return;
      }

      this.isLoading = true;

      if (this.isReacted) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/reactions/remove',
          data: {
            resource_type: this.resourceType,
            resource_id: this.resourceId,
            kind: 'thumbs_up',
          },
        })
          .done(() => {
            this.reactionsCount += -1;
            return (this.isReacted = false);
          })
          .always(() => {
            return (this.isLoading = false);
          });
      } else {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/reactions/add',
          data: {
            resource_type: this.resourceType,
            resource_id: this.resourceId,
            kind: 'thumbs_up',
            page_category: gon.page.category,
          },
        })
          .done(() => {
            this.reactionsCount += 1;
            return (this.isReacted = true);
          })
          .always(() => {
            return (this.isLoading = false);
          });
      }
    },
  },
};
