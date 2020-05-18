import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="text-center" v-if="!fetched">
      <div class="c-activity-group-more-button btn btn-secondary py-1 w-100" @click="more">
        <slot></slot>
      </div>
    </div>
  `,

  props: {
    username: {
      type: String,
      required: true,
    },

    pageCategory: {
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
        url: '/api/internal/activity_groups',
        data: {
          username: this.username,
          page_category: this.pageCategory,
          cursor: this.cursor,
        },
      }).done((html) => {
        eventHub.$emit('activity-group-more:fetched', this.cursor, html);
        this.fetched = true;
      });
    },
  },
};
