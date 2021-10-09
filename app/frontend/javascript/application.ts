import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';
import 'bootstrap/js/dist/modal';
import Popover from 'bootstrap/js/dist/popover';
import 'dayjs/locale/ja';

import { Application } from "@hotwired/stimulus"
import * as Turbo from '@hotwired/turbo';
import { definitionsFromContext } from "@hotwired/stimulus-webpack-helpers"
import axios from 'axios';
import ujs from '@rails/ujs';
import dayjs from 'dayjs';
import Cookies from 'js-cookie';

import { getTimeZone } from './utils/time-zone';

const annConfig = (window as any).AnnConfig;

document.addEventListener('turbo:load', (_event) => {
  if (typeof gtag == 'function') {
    gtag('js', new Date());
    gtag('config', annConfig.ga.trackingId);
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

const application = Application.start();
const context = (require as any).context('./controllers', true, /\.ts$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbo.start();
