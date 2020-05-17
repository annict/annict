import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="mb-3 text-center" v-if="!fetched">
      <div class="c-activity-group-more-button btn btn-secondary py-1 w-100" @click="more">
        <slot></slot>
      </div>
    </div>
  `,

  props: {
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
        url: '/api/internal/activity_groups',
        data: {
          cursor: this.cursor,
        },
      }).done((html) => {
        eventHub.$emit('activity-group-more:fetched', this.cursor, html);
        this.fetched = true;
      });
    },
  },
};
