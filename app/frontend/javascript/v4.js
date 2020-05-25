import 'bootstrap/js/dist/collapse';
import 'bootstrap/js/dist/dropdown';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';

import vueApp from './web/vueApp';
import lazyLoad from './web/utils/lazyLoad';

document.addEventListener('turbolinks:load', (_event) => {
  vueApp.start();
  lazyLoad.update();
});

ujs.start();
Turbolinks.start();
