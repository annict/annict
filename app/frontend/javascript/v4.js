import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';

import ujs from '@rails/ujs';
import $ from "jquery";
import Cookies from "js-cookie";
import moment from "moment-timezone";
import Turbolinks from 'turbolinks';

import vueApp from './web/vueApp';
import lazyLoad from './web/utils/lazyLoad';

document.addEventListener('turbolinks:load', (_event) => {
  moment.locale(AnnConfig.viewer.locale);

  Cookies.set('ann_time_zone', moment.tz.guess(), {
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
