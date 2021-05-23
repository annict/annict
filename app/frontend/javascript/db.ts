import * as Turbo from '@hotwired/turbo';
import ujs from '@rails/ujs';
import { Application } from 'stimulus';
import { definitionsFromContext } from 'stimulus/webpack-helpers';

const application = Application.start();
const context = (require as any).context('./db/controllers', true, /\.ts$/);
application.load(definitionsFromContext(context));

ujs.start();
Turbo.start();
