import 'bootstrap';
import 'select2';
import ujs from '@rails/ujs';
import Turbolinks from 'turbolinks';

document.addEventListener('turbolinks:load', _event => {
  WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

ujs.start();
Turbolinks.start();
