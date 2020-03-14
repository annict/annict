import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '#t-channel-receive-button',

  props: {
    channelId: {
      type: Number,
      required: true,
    },
    initIsReceiving: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      isReceiving: this.initIsReceiving,
      isSaving: false,
    };
  },

  methods: {
    toggle() {
      this.isSaving = true;

      if (this.isReceiving) {
        return $.ajax({
          method: 'DELETE',
          url: `/api/internal/receptions/${this.channelId}`,
        }).done(() => {
          this.isReceiving = false;
          return (this.isSaving = false);
        });
      } else {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/receptions',
          data: {
            channel_id: this.channelId,
          },
        }).done(() => {
          this.isReceiving = true;
          return (this.isSaving = false);
        });
      }
    },
  },
};
