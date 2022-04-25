import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';
import 'bootstrap/js/dist/modal';
import Popover from 'bootstrap/js/dist/popover';
import 'dayjs/locale/ja';

import { Application } from "@hotwired/stimulus"
import * as Turbo from '@hotwired/turbo';
import axios from 'axios';
import ujs from '@rails/ujs';
import dayjs from 'dayjs';
import Cookies from 'js-cookie';

import BasicFormController from './controllers/forms/basic-form-controller'
import BodyController from './controllers/body-controller'
import BulkWatchEpisodesButtonController from './controllers/bulk-watch-episodes-button-controller'
import CharactersCounterController from './controllers/characters-counter-controller'
import ComponentValueFetcherController from './controllers/component-value-fetcher-controller'
import EpisodeRecordFormController from './controllers/forms/episode-record-form-controller'
import FlashController from './controllers/flash-controller'
import FollowButtonController from './controllers/follow-button-controller'
import LikeButtonController from './controllers/like-button-controller'
import MainSidebarController from './controllers/main-sidebar-controller'
import MuteUserButtonController from './controllers/mute-user-button-controller'
import ProgramSelectRadioController from './controllers/program-select-radio-controller'
import ReceiveChannelButtonController from './controllers/receive-channel-button-controller'
import RecordRatingController from './controllers/record-rating-controller'
import RecordTextareaController from './controllers/record-textarea-controller'
import RelativeTimeController from './controllers/relative-time-controller'
import ReloadableController from './controllers/reloadable-controller'
import ResourceSelectorController from './controllers/resource-selector-controller'
import ShareToFacebookButtonController from './controllers/share-to-facebook-button-controller'
import ShareToTwitterButtonController from './controllers/share-to-twitter-button-controller'
import SkipEpisodeButtonController from './controllers/skip-episode-button-controller'
import SpoilerGuardController from './controllers/spoiler-guard-controller'
import StarButtonController from './controllers/star-button-controller'
import StatusSelectDropdownController from './controllers/status-select-dropdown-controller'
import TabBarController from './controllers/tab-bar-controller'
import TrackingHeatmapController from './controllers/tracking-heatmap-controller'
import TrackingOffcanvasButtonController from './controllers/tracking-offcanvas-button-controller'
import TrackingOffcanvasController from './controllers/tracking-offcanvas-controller'
import WatchEpisodeButtonController from './controllers/watch-episode-button-controller'
import WorkRecordFormController from './controllers/forms/work-record-form-controller'

import { getTimeZone } from './utils/time-zone';

const annConfig = (window as any).AnnConfig;

document.addEventListener('turbo:load', (_event) => {
  if (typeof gtag == 'function') {
    gtag('js', new Date());
    gtag('config', annConfig.ga.trackingId);
  }

  const ads = document.querySelectorAll('.adsbygoogle');
  if (ads.length > 0) {
    ads.forEach(function (ad) {
      if (ad.firstChild) {
        ad.removeChild(ad.firstChild);
      }
      (window as any).adsbygoogle = (window as any).adsbygoogle || [];
      (window as any).adsbygoogle.push({});
    });
  }

  Cookies.set('ann_time_zone', getTimeZone(), {
    domain: `.${annConfig.domain}`,
    secure: annConfig.rails.env === 'production',
  });

  axios.defaults.headers.common['X-CSRF-Token'] = document
    .querySelector('meta[name="csrf-token"]')
    ?.getAttribute('content');

  const popoverTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="popover"]'));
  const popoverList = popoverTriggerList.map(function (popoverTriggerEl) {
    return new Popover(popoverTriggerEl);
  });
});

dayjs.locale(annConfig.viewer.locale);

window.Stimulus = Application.start();
Stimulus.register('forms--basic-form', BasicFormController);
Stimulus.register('body', BodyController);
Stimulus.register('bulk-watch-episodes-button', BulkWatchEpisodesButtonController);
Stimulus.register('characters-counter', CharactersCounterController);
Stimulus.register('component-value-fetcher', ComponentValueFetcherController);
Stimulus.register('forms--episode-record-form', EpisodeRecordFormController);
Stimulus.register('flash', FlashController);
Stimulus.register('follow-button', FollowButtonController);
Stimulus.register('like-button', LikeButtonController);
Stimulus.register('main-sidebar', MainSidebarController);
Stimulus.register('mute-user-button', MuteUserButtonController);
Stimulus.register('program-select-radio', ProgramSelectRadioController);
Stimulus.register('receive-channel-button', ReceiveChannelButtonController);
Stimulus.register('record-rating', RecordRatingController);
Stimulus.register('record-textarea', RecordTextareaController);
Stimulus.register('relative-time', RelativeTimeController);
Stimulus.register('reloadable', ReloadableController);
Stimulus.register('resource-selector', ResourceSelectorController);
Stimulus.register('share-to-facebook-button', ShareToFacebookButtonController);
Stimulus.register('share-to-twitter-button', ShareToTwitterButtonController);
Stimulus.register('skip-episode-button', SkipEpisodeButtonController);
Stimulus.register('spoiler-guard', SpoilerGuardController);
Stimulus.register('star-button', StarButtonController);
Stimulus.register('status-select-dropdown', StatusSelectDropdownController);
Stimulus.register('tab-bar', TabBarController);
Stimulus.register('tracking-heatmap', TrackingHeatmapController);
Stimulus.register('tracking-offcanvas-button', TrackingOffcanvasButtonController);
Stimulus.register('tracking-offcanvas', TrackingOffcanvasController);
Stimulus.register('watch-episode-button', WatchEpisodeButtonController);
Stimulus.register('forms--work-record-form', WorkRecordFormController);

ujs.start();
Turbo.start();
