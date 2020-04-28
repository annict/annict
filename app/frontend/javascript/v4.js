import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';
import LazyLoad from 'vanilla-lazyload';
import Vue from 'vue';

import content from './web/components/content';
import shareButtonFacebook from './web/components/shareButtonFacebook';
import shareButtonTwitter from './web/components/shareButtonTwitter';
import sidebar from './web/components/sidebar';
import statusSelector from './web/components/statusSelector';
import tabBar from './web/components/tabBar';

document.addEventListener('turbolinks:load', (_event) => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });

  new Vue({
    el: '.ann-application',
    components: {
      'c-content': content,
      'c-share-to-facebook-button': shareButtonFacebook,
      'c-share-to-twitter-button': shareButtonTwitter,
      'c-sidebar': sidebar,
      'c-status-selector': statusSelector,
      'c-tab-bar': tabBar,
    },
  });

  new LazyLoad({
    elements_selector: '.js-lazy',
  });
});

ujs.start();
Turbolinks.start();
