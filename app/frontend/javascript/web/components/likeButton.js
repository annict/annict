import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: `
    <div :class="{ 'c-like-button': true, 'd-inline-block': true, 'u-fake-link': true, 'is-liked': isLiked }" @click="toggleLike">
      <i :class="{ 'far fa-heart': !isLiked, 'fas fa-heart': isLiked }"></i>
      <span class="count">
        {{ likesCount }}
      </span>
    </div>
  `,

  props: {
    resourceName: {
      type: String,
      required: true,
    },
    resourceId: {
      type: Number,
      required: true,
    },
    initLikesCount: {
      type: Number,
      required: true,
    },
    isSignedIn: {
      type: Boolean,
      default: false,
    },
    pageCategory: {
      type: String,
    },
  },

  data() {
    return {
      likesCount: Number(this.initLikesCount),
      isLiked: false,
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
        $.ajax({
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
        $.ajax({
          method: 'POST',
          url: '/api/internal/likes',
          data: {
            recipient_type: this.resourceName,
            recipient_id: this.resourceId,
            page_category: this.pageCategory,
          },
        }).done(() => {
          this.isLoading = false;
          this.likesCount += 1;
          this.isLiked = true;
        });
      }
    },
  },

  mounted() {
    eventHub.$on('request:likes:fetched', (likes) => {
      const like = likes.filter((like) => {
        return like.recipient_type === this.resourceName && like.recipient_id === this.resourceId;
      })[0];

      this.isLiked = !!like;
    });
  },
};
