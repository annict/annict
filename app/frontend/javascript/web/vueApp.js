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
    new Vue({
      el: '.ann-application',

      components: {
        'c-activity-group-more-button': activityGroupMoreButton,
        'c-activity-group-more-content': activityGroupMoreContent,
        'c-activity-more-button': activityMoreButton,
        'c-activity-more-content': activityMoreContent,
        'c-flash': flash,
        'c-follow-button': followButton,
        'c-like-button': likeButton,
        'c-rating-label': ratingLabel,
        'c-relative-time': relativeTime,
        'c-share-to-facebook-button': shareToFacebookButton,
        'c-share-to-twitter-button': shareToTwitterButton,
        'c-sidebar': sidebar,
        'c-sign-up-modal': signUpModal,
        'c-spoiler-guard': spoilerGuard,
        'c-status-selector': statusSelector,
        'c-tab-bar': tabBar,
        'c-user-actions-dropdown': userActionsDropdown,
        'c-user-data-fetcher': userDataFetcher,
        'c-user-heatmap': userHeatmap,
        'c-work-status-chart': workStatusChart,
        'c-work-watchers-chart': workWatchersChart,
      },
    });
  },
};
