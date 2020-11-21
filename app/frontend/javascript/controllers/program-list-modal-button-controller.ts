import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  openModal() {
    const animeId = this.data.get('animeId')
    new EventDispatcher('program-list-modal:open', { animeId }).dispatch();
  }
}
