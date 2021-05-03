import * as Turbo from '@hotwired/turbo';
import ujs from '@rails/ujs';
import { Application } from 'stimulus';
import { definitionsFromContext } from 'stimulus/webpack-helpers';

document.addEventListener('turbo:load', (_event) => {
  (window as any).WebFont.load({
    google: {
      families: ['Raleway'],
    },
  });
});

const application = Application.start();
const context = (require as any).context('./controllers', true, /\.ts$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbo.start();
