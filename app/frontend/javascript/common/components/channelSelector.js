import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '#t-channel-selector',

  props: {
    workId: {
      type: Number,
      required: true,
    },

    initChannelId: {
      type: String,
      required: true,
    },

    options: {
      type: Array,
      required: true,
    },
  },

  data() {
    return {
      channelId: this.initChannelId,
      isSaving: false,
    };
  },

  methods: {
    change() {
      this.isSaving = true;

      return $.ajax({
        method: 'POST',
        url: `/api/internal/works/${this.workId}/channels/select`,
        data: {
          channel_id: this.channelId,
        },
      }).done(() => {
        return (this.isSaving = false);
      });
    },
  },
};
