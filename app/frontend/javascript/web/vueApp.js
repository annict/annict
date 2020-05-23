import Vue from 'vue';

import activityGroupMoreButton from './components/activityGroupMoreButton';
import activityGroupMoreContent from './components/activityGroupMoreContent';
import activityMoreButton from './components/activityMoreButton';
import activityMoreContent from './components/activityMoreContent';
import flash from './components/flash';
import followButton from './components/followButton';
import likeButton from './components/likeButton';
import ratingLabel from './components/ratingLabel';
import relativeTime from './components/relativeTime';
import shareToFacebookButton from './components/shareToFacebookButton';
import shareToTwitterButton from './components/shareToTwitterButton';
import sidebar from './components/sidebar';
import signUpModal from './components/signUpModal';
import spoilerGuard from './components/spoilerGuard';
import statusSelector from './components/statusSelector';
import tabBar from './components/tabBar';
import userActionsDropdown from './components/userActionsDropdown';
import userDataFetcher from './components/userDataFetcher';
import userHeatmap from './components/userHeatmap';
import workStatusChart from './components/workStatusChart';
import workWatchersChart from './components/workWatchersChart';

export default {
  start() {
    Vue.component('c-activity-group-more-button', activityGroupMoreButton);
    Vue.component('c-activity-group-more-content', activityGroupMoreContent);
    Vue.component('c-activity-more-button', activityMoreButton);
    Vue.component('c-activity-more-content', activityMoreContent);
    Vue.component('c-flash', flash);
    Vue.component('c-follow-button', followButton);
    Vue.component('c-like-button', likeButton);
    Vue.component('c-rating-label', ratingLabel);
    Vue.component('c-relative-time', relativeTime);
    Vue.component('c-share-to-facebook-button', shareToFacebookButton);
    Vue.component('c-share-to-twitter-button', shareToTwitterButton);
    Vue.component('c-sidebar', sidebar);
    Vue.component('c-sign-up-modal', signUpModal);
    Vue.component('c-spoiler-guard', spoilerGuard);
    Vue.component('c-status-selector', statusSelector);
    Vue.component('c-tab-bar', tabBar);
    Vue.component('c-user-actions-dropdown', userActionsDropdown);
    Vue.component('c-user-data-fetcher', userDataFetcher);
    Vue.component('c-user-heatmap', userHeatmap);
    Vue.component('c-work-status-chart', workStatusChart);
    Vue.component('c-work-watchers-chart', workWatchersChart);

    new Vue({
      el: '.ann-application',
    });
  },
};
