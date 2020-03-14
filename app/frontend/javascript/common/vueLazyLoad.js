import $ from 'jquery';

export default {
  refresh() {
    // Scroll 1px to load images
    return $(window).scrollTop($(window).scrollTop() + 1);
  },
};
