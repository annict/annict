import axios from 'axios';

import vueApp from './common/vueApp';

document.addEventListener('turbo:load', (_event) => {
  axios.defaults.headers.common['X-CSRF-Token'] = document
    .querySelector('meta[name="csrf-token"]')
    ?.getAttribute('content');

  vueApp.start();
});
