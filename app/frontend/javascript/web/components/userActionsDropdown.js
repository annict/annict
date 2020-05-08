import $ from 'jquery';
import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="c-user-actions-dropdown dropdown d-inline-block">
      <div class="btn btn-outline-secondary dropdown-toggle" data-toggle="dropdown">
        <div class="dropdown-menu dropdown-menu-right">
          <a href="#" class="dropdown-item" @click.prevent="mute">
            {{ i18n.verb.mute }}
          </a>
        </div>
      </div>
    </div>
  `,

  props: {
    userId: {
      type: Number,
      required: true,
    },
  },

  computed: {
    i18n() {
      return AnnConfig.i18n;
    },
  },

  methods: {
    mute() {
      if (confirm(this.i18n.messages.areYouSure)) {
        $.ajax({
          method: 'POST',
          url: '/api/internal/mute_users',
          data: {
            user_id: this.userId,
          },
        }).done(() => {
          eventHub.$emit('muteUser:mute', this.userId);
          const msg = this.i18n.messages.userHasBeenMuted;
          eventHub.$emit('flash:show', msg);
        });
      }
    },
  },
};
