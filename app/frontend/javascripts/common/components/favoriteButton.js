import $ from 'jquery';

export default {
  template: '#t-favorite-button',

  props: {
    resourceType: {
      type: String,
      required: true,
    },
    resourceId: {
      type: Number,
      required: true,
    },
    initIsFavorited: {
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
      isFavorited: this.initIsFavorited,
      isSaving: false,
    };
  },

  computed: {
    buttonText() {
      if (this.isFavorited) {
        return gon.I18n['messages._components.favorite_button.added_to_favorites'];
      } else {
        return gon.I18n['messages._components.favorite_button.add_to_favorites'];
      }
    },
  },

  methods: {
    toggleFavorite() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }

      this.isSaving = true;

      if (this.isFavorited) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/favorites/unfavorite',
          data: {
            resource_type: this.resourceType,
            resource_id: this.resourceId,
          },
        }).done(() => {
          this.isFavorited = false;
          return (this.isSaving = false);
        });
      } else {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/favorites',
          data: {
            resource_type: this.resourceType,
            resource_id: this.resourceId,
            page_category: gon.page.category,
          },
        }).done(() => {
          this.isFavorited = true;
          return (this.isSaving = false);
        });
      }
    },
  },
};
