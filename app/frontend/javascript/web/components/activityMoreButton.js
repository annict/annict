import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="mb-3 text-center" v-if="!fetched">
      <div class="c-activity-more-button btn btn-outline-secondary btn-small py-1" @click="more">
        <slot></slot>
      </div>
    </div>
  `,

  props: {
    activityGroupId: {
      type: String,
      required: true,
    },
    cursor: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      fetched: false,
    };
  },

  methods: {
    more() {
      $.ajax({
        method: 'GET',
        url: '/api/internal/activities',
        data: {
          activity_group_id: this.activityGroupId,
          cursor: this.cursor,
        },
      }).done((html) => {
        eventHub.$emit('activity-more:fetched', this.activityGroupId, html);
        this.fetched = true;
      });
    },
  },
};
