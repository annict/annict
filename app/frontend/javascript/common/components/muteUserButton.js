import $ from 'jquery';
import Vue from 'vue';

import eventHub from '../../common/eventHub';
import { EventDispatcher } from '../../utils/event-dispatcher';

export default {
  template: '#t-mute-user-button',

  props: {
    userId: {
      type: Number,
      required: true,
    },
  },

  methods: {
    mute() {
      if (confirm(gon.I18n['messages._common.are_you_sure'])) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/mute_users',
          data: {
            user_id: this.userId,
          },
        }).done(() => {
          eventHub.$emit('muteUser:mute', this.userId);
          const message = gon.I18n['messages.components.mute_user_button.the_user_has_been_muted'];
          new EventDispatcher('flash:show', { message }).dispatch();
        });
      }
    },
  },
};
