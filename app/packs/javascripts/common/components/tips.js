import $ from 'jquery';

export default {
  template: '#t-tips',

  data() {
    return { tips: JSON.parse(gon.tips) };
  },

  methods: {
    open(index) {
      const tip = this.tips[index];
      return (tip.isOpened = !tip.isOpened);
    },

    close(index) {
      if (confirm(gon.I18n['messages._common.are_you_sure'])) {
        return $.ajax({
          method: 'POST',
          url: '/api/internal/tips/close',
          data: {
            slug: this.tips[index].slug,
            page_category: gon.basic.pageCategory
          }
        }).done(() => {
          return this.tips.splice(index, 1);
        });
      }
    }
  }
};
