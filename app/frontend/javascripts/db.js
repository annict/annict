import 'bootstrap';
import 'select2';
import ujs from '@rails/ujs';
import { Application } from 'stimulus';
import { definitionsFromContext } from 'stimulus/webpack-helpers';
import Turbolinks from 'turbolinks';

document.addEventListener('turbolinks:load', _event => {
  window.WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

const application = Application.start();
const context = require.context('./db/controllers', true, /\.js$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbolinks.start();
