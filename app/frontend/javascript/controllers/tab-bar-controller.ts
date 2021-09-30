import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  showSidebar(event: Event) {
    event.preventDefault();
    new EventDispatcher('main-sidebar:show').dispatch();
  }
}
