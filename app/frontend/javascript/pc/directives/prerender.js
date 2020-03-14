import $ from 'jquery';

export default {
  inserted(el) {
    $(el).on('mouseenter', () => {
      const url = $(el).prop('href');
      let $prerender = $('link[rel="prerender"]');
      if ($prerender.length) {
        $prerender.prop('href', url);
      } else {
        $('<link>')
          .prop('rel', 'prerender')
          .prop('href', url)
          .appendTo('head');
      }
    });
  },
};
