import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';
import Vue from 'vue';

import ResourceSelector from './db/components/ResourceSelector';

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });

  Vue.component('c-resource-selector', ResourceSelector);

  new Vue({
    el: '.f-app',
  });
});

ujs.start();
Turbolinks.start();
