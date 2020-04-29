import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';
import LazyLoad from 'vanilla-lazyload';
import Vue from 'vue';

import content from './web/components/content';
import likeButton from './web/components/likeButton';
import ratingLabel from './web/components/ratingLabel';
import shareToFacebookButton from './web/components/shareToFacebookButton';
import shareToTwitterButton from './web/components/shareToTwitterButton';
import sidebar from './web/components/sidebar';
import statusSelector from './web/components/statusSelector';
import tabBar from './web/components/tabBar';
import workStatusChart from './web/components/workStatusChart';
import workWatchersChart from './web/components/workWatchersChart';

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
      'c-like-button': likeButton,
      'c-rating-label': ratingLabel,
      'c-share-to-facebook-button': shareToFacebookButton,
      'c-share-to-twitter-button': shareToTwitterButton,
      'c-sidebar': sidebar,
      'c-status-selector': statusSelector,
      'c-tab-bar': tabBar,
      'c-work-status-chart': workStatusChart,
      'c-work-watchers-chart': workWatchersChart,
    },
  });

  new LazyLoad({
    elements_selector: '.js-lazy',
  });
});

ujs.start();
Turbolinks.start();
