import $ from 'jquery';
import eventHub from '../../common/eventHub';

export default {
  template: `
    <div
      class="c-follow-button btn"
      :class="{ 'u-btn-outline-green': !isFollowing, 'u-btn-green': isFollowing, 'c-spinner': isSaving }"
      @click="toggle"
    >
      <i class="fa mr-2" :class="{ 'fa-plus': !isFollowing, 'fa-check': isFollowing }"></i>
      <span>
        {{ buttonText }}
      </span>
    </div>
  `,

  props: {
    userId: {
      type: Number,
      required: true,
    },
  },

  data() {
    return {
      isFollowing: false,
      isSaving: false,
    };
  },

  computed: {
    i18n() {
      return AnnConfig.i18n;
    },

    buttonText() {
      if (this.isFollowing) {
        return this.i18n.noun.following;
      } else {
        return this.i18n.verb.follow;
      }
    },
  },

  methods: {
    toggle() {
      this.isSaving = true;

      if (this.isFollowing) {
        $.ajax({
          method: 'POST',
          url: '/api/internal/follows/unfollow',
          data: {
            user_id: this.userId,
          },
        }).done(() => {
          this.isFollowing = false;
          this.isSaving = false;
        });
      } else {
        $.ajax({
          method: 'POST',
          url: '/api/internal/follows',
          data: {
            user_id: this.userId,
          },
        })
          .done(() => {
            this.isFollowing = true;
            this.isSaving = false;
          })
          .fail(() => {
            $('.c-sign-up-modal').modal('show');
            this.isSaving = false;
          });
      }
    },
  },

  mounted() {
    eventHub.$on('request:following:fetched', (following) => {
      this.isFollowing = !!following.filter((f) => {
        return f.user_id === this.userId;
      })[0];
    });
  },
};
