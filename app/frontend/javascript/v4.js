import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';

import 'dayjs/locale/ja'

import ujs from '@rails/ujs';
import dayjs from 'dayjs'
import $ from "jquery";
import Cookies from "js-cookie";
import Turbolinks from 'turbolinks';

import vueApp from './web/vueApp';
import lazyLoad from './web/utils/lazyLoad';
import { getTimeZone } from './web/utils/timeZone'

document.addEventListener('turbolinks:load', (_event) => {
  dayjs.locale(AnnConfig.viewer.locale);

  Cookies.set('ann_time_zone', getTimeZone(), {
    domain: `.${AnnConfig.domain}`,
    secure: AnnConfig.rails.env === 'production',
  });

  $.ajaxSetup({
    headers: { 'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content') },
  });

  vueApp.start();
  lazyLoad.update();
});

ujs.start();
Turbolinks.start();
