import $ from 'jquery';

export default {
  template: '#t-follow-button',

  props: {
    username: {
      type: String,
      required: true,
    },
    initIsFollowing: {
      type: Boolean,
      required: true,
    },
    isSmall: {
      type: Boolean,
      default: false,
    },
    isSignedIn: {
      type: Boolean,
      default: false,
    },
  },

  data() {
    return {
      isFollowing: this.initIsFollowing,
      isSaving: false,
    };
  },

  computed: {
    buttonText() {
      if (this.isFollowing) {
        return window.gon.I18n['noun.following'];
      } else {
        return window.gon.I18n['verb.follow'];
      }
    },
  },

  methods: {
    toggle() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }

      this.isSaving = true;

      if (this.isFollowing) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/follows/unfollow',
          data: {
            username: this.username,
          },
        }).done(() => {
          this.isFollowing = false;
          return (this.isSaving = false);
        });
      } else {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/follows',
          data: {
            username: this.username,
            page_category: gon.page.category,
          },
        }).done(() => {
          this.isFollowing = true;
          return (this.isSaving = false);
        });
      }
    },
  },
};
