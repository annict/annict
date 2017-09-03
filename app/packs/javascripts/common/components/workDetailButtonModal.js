import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: '#t-work-detail-button-modal',

  data() {
    return {
      work: null,
      status: null,
      isLoadingWorkDetail: false,
      workId: null
    };
  },

  methods: {
    loadWorkDetail() {
      return $.ajax({
        method: 'GET',
        url: `/api/internal/works/${this.workId}`
      })
        .done(data => {
          this.work = data.work;
          this.status = data.status;
          return (this.workImageUrl = data.work.image_url);
        })
        .fail(function() {
          const message =
            gon.I18n['messages._components.work_detail_button.error'];
          return eventHub.$emit('flash:show', message, 'alert');
        })
        .always(() => {
          return (this.isLoadingWorkDetail = false);
        });
    }
  },

  created() {
    return eventHub.$on('workDetailButtonModal:show', workId => {
      this.workId = workId;
      this.isLoadingWorkDetail = true;
      $('.c-work-detail-button-modal').modal('show');
      return this.loadWorkDetail();
    });
  }
};
