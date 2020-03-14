import $ from 'jquery';

export default {
  template: '#t-tips',

  data() {
    return {
      tips: JSON.parse(window.gon.tipsData),
    };
  },

  methods: {
    close(index) {
      if (confirm(window.gon.I18n['messages._common.are_you_sure'])) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/tips/close',
          data: {
            slug: this.tips[index].slug,
            page_category: window.gon.page.category,
          },
        }).done(() => {
          return this.tips.splice(index, 1);
        });
      }
    },
  },
};
