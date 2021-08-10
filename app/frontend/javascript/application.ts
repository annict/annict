import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';
import 'bootstrap/js/dist/modal';
import 'dayjs/locale/ja';

import * as Turbo from '@hotwired/turbo';
import axios from 'axios';
import ujs from '@rails/ujs';
import dayjs from 'dayjs';
import Cookies from 'js-cookie';
import { Application } from 'stimulus';
import { definitionsFromContext } from 'stimulus/webpack-helpers';

import { getTimeZone } from './utils/time-zone';

document.addEventListener('turbo:load', (_event) => {
  const annConfig = (window as any).AnnConfig;

  if (typeof gtag == 'function') {
    gtag('js', new Date());
    gtag('config', annConfig.ga.trackingId);
  }

  dayjs.locale(annConfig.viewer.locale);

  Cookies.set('ann_time_zone', getTimeZone(), {
    domain: `.${annConfig.domain}`,
    secure: annConfig.rails.env === 'production',
  });

  axios.defaults.headers.common['X-CSRF-Token'] = document
    .querySelector('meta[name="csrf-token"]')
    ?.getAttribute('content');
});

const application = Application.start();
const context = (require as any).context('./controllers', true, /\.ts$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbo.start();
