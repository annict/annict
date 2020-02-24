import 'bootstrap';
import railsUjs from '@rails/ujs';
import Turbolinks from 'turbolinks';

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

railsUjs.start();
Turbolinks.start();
