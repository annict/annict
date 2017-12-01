import $ from 'jquery';

import vueLazyLoad from '../../common/vueLazyLoad';

import createRecordActivity from './createRecordActivity';
import createReviewActivity from './createReviewActivity';
import createMultipleRecordsActivity from './createMultipleRecordsActivity';
import createStatusActivity from './createStatusActivity';
import loadMoreButton from './loadMoreButton';

export default {
  template: '#t-activities',

  props: {
    username: {
      type: String
    }
  },

  data() {
    return {
      isLoading: false,
      hasNext: false,
      activities: [],
      page: 1
    };
  },

  components: {
    'c-create-record-activity': createRecordActivity,
    'c-create-review-activity': createReviewActivity,
    'c-create-multiple-records-activity': createMultipleRecordsActivity,
    'c-create-status-activity': createStatusActivity,
    'c-load-more-button': loadMoreButton
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
        data: this.requestData()
      }).done(data => {
        this.isLoading = false;

        if (data.activities.length > 0) {
          this.hasNext = true;
          this.activities.push(...Array.from(data.activities || []));
        } else {
          this.hasNext = false;
        }

        return this.$nextTick(() => vueLazyLoad.refresh());
      });
    },

    _activityData() {
      if (!this.gon.activityData) {
        return {};
      }
      return JSON.parse(this.gon.activityData);
    }
  },

  mounted() {
    this.gon = window.gon;

    if (this.gon.user.device === 'pc') {
      $(this.$el).css({ maxHeight: window.innerHeight * 0.7 });
    }

    return this.load();
  }
};
