import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';
import LazyLoad from 'vanilla-lazyload';

import vueApp from './web/vueApp';

document.addEventListener('turbolinks:load', (_event) => {
  vueApp.start();

  new LazyLoad({
    elements_selector: '.js-lazy',
  });
});

ujs.start();
Turbolinks.start();
