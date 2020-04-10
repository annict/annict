import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';
import Vue from 'vue';

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });

  new Vue({
    el: '.f-app',
  });
});

ujs.start();
Turbolinks.start();
