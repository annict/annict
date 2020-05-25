import $ from 'jquery';
import 'bootstrap';
import ujs from '@rails/ujs';
import Cookies from 'js-cookie';
import moment from 'moment-timezone';
import 'moment/locale/ja';
import Turbolinks from 'turbolinks';
import LazyLoad from 'vanilla-lazyload';
import Vue from 'vue';

import app from './common/app';
import eventHub from './common/eventHub';
import vueLazyLoad from './common/vueLazyLoad';

import omittedSynopsis from './common/components/omittedSynopsis';
import privacyPolicyModal from './common/components/privacyPolicyModal';
import slotList from './common/components/slotList';
import reactionButton from './common/components/reactionButton';
import stickyMessage from './common/components/stickyMessage';
import usernamePreview from './common/components/usernamePreview';
import workComment from './common/components/workComment';
import workFriends from './common/components/workFriends';
import workTags from './common/components/workTags';
import youtubeModalPlayer from './common/components/youtubeModalPlayer';

import prerender from './pc/directives/prerender';

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

  Vue.component('c-privacy-policy-modal', privacyPolicyModal);
  Vue.component('c-reaction-button', reactionButton);
  Vue.component('c-slot-list', slotList);
  Vue.component('c-sticky-message', stickyMessage);
  Vue.component('c-username-preview', usernamePreview);
  Vue.component('c-work-comment', workComment);
  Vue.component('c-work-tags', workTags);

  Vue.directive('prerender', prerender);

  Vue.nextTick(() => {
    vueLazyLoad.refresh();
  });

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
