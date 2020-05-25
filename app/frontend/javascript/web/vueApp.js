import Vue from 'vue';
import VueLazyload from 'vue-lazyload';

import vueLazyLoad from '../common/vueLazyLoad';

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
import impressionButton from '../common/components/impressionButton';
import impressionButtonModal from '../common/components/impressionButtonModal';
import episodeList from "../common/components/episodeList";
import recordRating from "../common/components/recordRating";
import recordTextarea from "../common/components/recordTextarea";
import recordWordCount from "../common/components/recordWordCount";
import recordSorter from "../common/components/recordSorter";
import record from "../common/components/record";
import episodeRatingStateChart from "../common/components/episodeRatingStateChart";
import episodeRecordsChart from "../common/components/episodeRecordsChart";
import muteUserButton from "../common/components/muteUserButton";
import commentGuard from "../common/components/commentGuard";
import body from "../common/components/body";
import autosizeTextarea from "../common/components/autosizeTextarea";
import inputWordsCount from "../common/components/inputWordsCount";
import channelReceiveButton from "../common/components/channelReceiveButton";
import channelSelector from "../common/components/channelSelector";
import untrackedEpisodeList from "../common/components/untrackedEpisodeList";
import episodeProgress from "../common/components/episodeProgress";
import favoriteButton from "../common/components/favoriteButton";
import forumEditLink from "../common/components/forumEditLink";
import omittedSynopsis from "../common/components/omittedSynopsis";
import workFriends from "../common/components/workFriends";
import youtubeModalPlayer from "../common/components/youtubeModalPlayer";

export default {
  start() {
    Vue.use(VueLazyload);

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
    Vue.component('c-impression-button', impressionButton);
    Vue.component('c-impression-button-modal', impressionButtonModal);
    Vue.component('c-episode-list', episodeList);
    Vue.component('c-record-rating', recordRating);
    Vue.component('c-record-textarea', recordTextarea);
    Vue.component('c-record-word-count', recordWordCount);
    Vue.component('c-record-sorter', recordSorter);
    Vue.component('c-record', record);
    Vue.component('c-episode-rating-state-chart', episodeRatingStateChart);
    Vue.component('c-episode-records-chart', episodeRecordsChart);
    Vue.component('c-mute-user-button', muteUserButton);
    Vue.component('c-comment-guard', commentGuard);
    Vue.component('c-body', body);
    Vue.component('c-autosize-textarea', autosizeTextarea);
    Vue.component('c-input-words-count', inputWordsCount);
    Vue.component('c-channel-receive-button', channelReceiveButton);
    Vue.component('c-channel-selector', channelSelector);
    Vue.component('c-untracked-episode-list', untrackedEpisodeList);
    Vue.component('c-episode-progress', episodeProgress);
    Vue.component('c-favorite-button', favoriteButton);
    Vue.component('c-omitted-synopsis', omittedSynopsis);
    Vue.component('c-work-friends', workFriends);
    Vue.component('c-youtube-modal-player', youtubeModalPlayer);

    Vue.nextTick(() => {
      vueLazyLoad.refresh();
    });

    new Vue({
      el: '.ann-application',
    });
  },
};
