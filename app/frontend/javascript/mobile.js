import $ from 'jquery';
import 'bootstrap';
import 'select2';
import ujs from '@rails/ujs';
import Cookies from 'js-cookie';
import moment from 'moment-timezone';
import 'moment/locale/ja';
import Turbolinks from 'turbolinks';
import LazyLoad from 'vanilla-lazyload';
import Vue from 'vue';
import VueLazyload from 'vue-lazyload';

import app from './common/app';
import eventHub from './common/eventHub';
import vueLazyLoad from './common/vueLazyLoad';

import activities from './common/components/activities';
import analytics from './common/components/analytics';
import autosizeTextarea from './common/components/autosizeTextarea';
import body from './common/components/body';
import channelReceiveButton from './common/components/channelReceiveButton';
import channelSelector from './common/components/channelSelector';
import commentGuard from './common/components/commentGuard';
import episodeList from './common/components/episodeList';
import episodeProgress from './common/components/episodeProgress';
import favoriteButton from './common/components/favoriteButton';
import forumEditLink from './common/components/forumEditLink';
import impressionButton from './common/components/impressionButton';
import impressionButtonModal from './common/components/impressionButtonModal';
import inputWordsCount from './common/components/inputWordsCount';
import omittedSynopsis from './common/components/omittedSynopsis';
import muteUserButton from './common/components/muteUserButton';
import privacyPolicyModal from './common/components/privacyPolicyModal';
import slotList from './common/components/slotList';
import reactionButton from './common/components/reactionButton';
import record from './common/components/record';
import recordRating from './common/components/recordRating';
import recordSorter from './common/components/recordSorter';
import recordTextarea from './common/components/recordTextarea';
import recordWordCount from './common/components/recordWordCount';
import timeAgo from './common/components/timeAgo';
import tips from './common/components/tips';
import untrackedEpisodeList from './common/components/untrackedEpisodeList';
import userHeatmap from './common/components/userHeatmap';
import usernamePreview from './common/components/usernamePreview';
import workComment from './common/components/workComment';
import workFriends from './common/components/workFriends';
import workTags from './common/components/workTags';
import youtubeModalPlayer from './common/components/youtubeModalPlayer';

import flash from './web/components/flash';
import followButton from './web/components/followButton';
import likeButton from './web/components/likeButton';
import ratingLabel from './web/components/ratingLabel';
import shareToTwitterButton from './web/components/shareToTwitterButton';
import shareToFacebookButton from './web/components/shareToFacebookButton';
import sidebar from './web/components/sidebar';
import statusSelector from './web/components/statusSelector';
import tabBar from './web/components/tabBar';
import userDataFetcher from './web/components/userDataFetcher';

import resourceSelect from './common/directives/resourceSelect';

document.addEventListener('turbolinks:load', (event) => {
  moment.locale(gon.user.locale);
  Cookies.set('ann_time_zone', moment.tz.guess(), {
    domain: `.${gon.annict.domain}`,
    secure: gon.rails.env === 'production',
  });

  $.ajaxSetup({
    headers: { 'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content') },
  });

  WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });

  Vue.use(VueLazyload);

  Vue.component('c-activities', activities);
  Vue.component('c-analytics', analytics(event));
  Vue.component('c-autosize-textarea', autosizeTextarea);
  Vue.component('c-body', body);
  Vue.component('c-channel-receive-button', channelReceiveButton);
  Vue.component('c-channel-selector', channelSelector);
  Vue.component('c-comment-guard', commentGuard);
  Vue.component('c-episode-list', episodeList);
  Vue.component('c-episode-progress', episodeProgress);
  Vue.component('c-favorite-button', favoriteButton);
  Vue.component('c-flash', flash);
  Vue.component('c-forum-edit-link', forumEditLink);
  Vue.component('c-follow-button', followButton);
  Vue.component('c-impression-button', impressionButton);
  Vue.component('c-impression-button-modal', impressionButtonModal);
  Vue.component('c-input-words-count', inputWordsCount);
  Vue.component('c-like-button', likeButton);
  Vue.component('c-omitted-synopsis', omittedSynopsis);
  Vue.component('c-mute-user-button', muteUserButton);
  Vue.component('c-privacy-policy-modal', privacyPolicyModal);
  Vue.component('c-slot-list', slotList);
  Vue.component('c-rating-label', ratingLabel);
  Vue.component('c-reaction-button', reactionButton);
  Vue.component('c-record', record);
  Vue.component('c-record-rating', recordRating);
  Vue.component('c-record-sorter', recordSorter);
  Vue.component('c-record-textarea', recordTextarea);
  Vue.component('c-record-word-count', recordWordCount);
  Vue.component('c-share-button-facebook', shareToFacebookButton);
  Vue.component('c-share-button-twitter', shareToTwitterButton);
  Vue.component('c-sidebar', sidebar);
  Vue.component('c-status-selector', statusSelector);
  Vue.component('c-tab-bar', tabBar);
  Vue.component('c-time-ago', timeAgo);
  Vue.component('c-tips', tips);
  Vue.component('c-untracked-episode-list', untrackedEpisodeList);
  Vue.component('c-user-data-fetcher', userDataFetcher);
  Vue.component('c-user-heatmap', userHeatmap);
  Vue.component('c-username-preview', usernamePreview);
  Vue.component('c-work-comment', workComment);
  Vue.component('c-work-friends', workFriends);
  Vue.component('c-work-tags', workTags);
  Vue.component('c-youtube-modal-player', youtubeModalPlayer);

  Vue.directive('resource-select', resourceSelect);

  Vue.nextTick(() => vueLazyLoad.refresh());

  new Vue({
    el: '.p-application',

    data() {
      return {
        appData: {},
        pageData: {},
      };
    },

    created() {
      app.loadAppData().done((appData) => {
        this.appData = appData;

        if (!appData.isUserSignedIn || !app.existsPageParams()) {
          eventHub.$emit('app:loaded');
          return;
        }

        app.loadPageData().done((pageData) => {
          this.pageData = pageData;
          eventHub.$emit('app:loaded');
        });
      });
    },
  });

  new LazyLoad({
    elements_selector: '.js-lazy',
  });
});

ujs.start();
Turbolinks.start();
