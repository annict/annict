import $ from 'jquery';

export default {
  template: '#t-record-sorter',

  props: {
    currentUrl: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      sort: gon.currentRecordsSortType,
      sortTypes: gon.recordsSortTypes,
    };
  },

  methods: {
    reload() {
      return this.updateRecordsSortType(() => {
        return (location.href = this.currentUrl);
      });
    },

    updateRecordsSortType(callback) {
      return $.ajax({
        method: 'PATCH',
        url: '/api/internal/records_sort_type',
        data: {
          records_sort_type: this.sort,
        },
      }).done(callback);
    },
  },
};
