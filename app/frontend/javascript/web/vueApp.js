import Vue from 'vue';

import likeButton from './components/likeButton';
import ratingLabel from './components/ratingLabel';
import shareToFacebookButton from './components/shareToFacebookButton';
import shareToTwitterButton from './components/shareToTwitterButton';
import sidebar from './components/sidebar';
import signUpModal from './components/signUpModal';
import spoilerGuard from './components/spoilerGuard';
import statusSelector from './components/statusSelector';
import tabBar from './components/tabBar';
import userDataFetcher from './components/userDataFetcher';
import workStatusChart from './components/workStatusChart';
import workWatchersChart from './components/workWatchersChart';

export default {
  start() {
    new Vue({
      el: '.ann-application',

      components: {
        'c-like-button': likeButton,
        'c-rating-label': ratingLabel,
        'c-share-to-facebook-button': shareToFacebookButton,
        'c-share-to-twitter-button': shareToTwitterButton,
        'c-sidebar': sidebar,
        'c-sign-up-modal': signUpModal,
        'c-spoiler-guard': spoilerGuard,
        'c-status-selector': statusSelector,
        'c-tab-bar': tabBar,
        'c-user-data-fetcher': userDataFetcher,
        'c-work-status-chart': workStatusChart,
        'c-work-watchers-chart': workWatchersChart,
      },
    });
  },
};
