import ujs from '@rails/ujs';
import { Application } from 'stimulus';
import { definitionsFromContext } from 'stimulus/webpack-helpers';
import Turbolinks from 'turbolinks';

document.addEventListener('turbolinks:load', (_event) => {
  (window as any).WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

const application = Application.start();
const context = (require as any).context('./db/controllers', true, /\.ts$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbolinks.start();
