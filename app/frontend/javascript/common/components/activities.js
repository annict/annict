import $ from 'jquery';

import vueLazyLoad from '../../common/vueLazyLoad';
import eventHub from '../eventHub';

import createRecordActivity from './createRecordActivity';
import createReviewActivity from './createReviewActivity';
import createMultipleRecordsActivity from './createMultipleRecordsActivity';
import createStatusActivity from './createStatusActivity';
import loadMoreButton from './loadMoreButton';

export default {
  template: '#t-activities',

  props: {
    username: {
      type: String,
    },
  },

  data() {
    return {
      isLoading: false,
      hasNext: false,
      activities: [],
      page: 1,
    };
  },

  components: {
    'c-create-episode-record-activity': createRecordActivity,
    'c-create-work-record-activity': createReviewActivity,
    'c-create-multiple-episode-records-activity': createMultipleRecordsActivity,
    'c-create-status-activity': createStatusActivity,
    'c-load-more-button': loadMoreButton,
  },

  methods: {
    requestData() {
      const data = { page: this.page };
      if (this.username) {
        data.username = this.username;
      }
      return data;
    },

    load() {
      this.isLoading = true;
      const { activities } = this._activityData();

      if (activities.length > 0) {
        this.hasNext = true;
        this.activities = activities;
      } else {
        this.hasNext = false;
      }

      this.isLoading = false;

      return this.$nextTick(() => vueLazyLoad.refresh());
    },

    loadMore() {
      this.isLoading = true;
      this.page += 1;

      return $.ajax({
        method: 'GET',
        url: '/api/internal/activities',
        data: this.requestData(),
      }).done((data) => {
        this.isLoading = false;

        if (data.activities.length > 0) {
          this.hasNext = true;
          this.activities.push(...Array.from(data.activities || []));
        } else {
          this.hasNext = false;
        }

        this.$nextTick(() => {
          vueLazyLoad.refresh();
          eventHub.$emit('content:refetch');
        });
      });
    },

    _activityData() {
      if (!gon.activityData) {
        return {};
      }
      return JSON.parse(gon.activityData);
    },
  },

  mounted() {
    if (gon.user.device === 'pc' && gon.page && gon.page.category === 'home_index') {
      $(this.$el).css({ maxHeight: window.innerHeight * 0.7 });
    }

    this.load();
  },
};
