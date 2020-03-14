import $ from 'jquery';

export default {
  template: '#t-like-button',

  props: {
    resourceName: {
      type: String,
      required: true,
    },
    initResourceId: {
      type: Number,
      required: true,
    },
    initLikesCount: {
      type: Number,
      required: true,
    },
    initIsLiked: {
      type: Boolean,
      required: true,
    },
    isSignedIn: {
      type: Boolean,
      default: false,
    },
  },

  data() {
    return {
      resourceId: Number(this.initResourceId),
      likesCount: Number(this.initLikesCount),
      isLiked: JSON.parse(this.initIsLiked),
      isLoading: false,
    };
  },

  methods: {
    toggleLike() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }

      if (this.isLoading) {
        return;
      }

      this.isLoading = true;

      if (this.isLiked) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/likes/unlike',
          data: {
            recipient_type: this.resourceName,
            recipient_id: this.resourceId,
          },
        }).done(() => {
          this.isLoading = false;
          this.likesCount += -1;
          this.isLiked = false;
        });
      } else {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/likes',
          data: {
            recipient_type: this.resourceName,
            recipient_id: this.resourceId,
            page_category: gon.page.category,
          },
        }).done(() => {
          this.isLoading = false;
          this.likesCount += 1;
          this.isLiked = true;
        });
      }
    },
  },
};
