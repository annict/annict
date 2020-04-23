import 'bootstrap';

import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

ujs.start();
Turbolinks.start();
