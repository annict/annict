import autosize from 'autosize';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  charactersCounterId!: string

  initialize() {
    this.charactersCounterId = this.data.get('charactersCounterId') ?? 'id';

    autosize(this.element)
  }

  updateCharactersCount(event: Event) {
    const charactersCount = event.currentTarget.value.length

    new EventDispatcher('characters-counter:update', { charactersCounterId: this.charactersCounterId, charactersCount }).dispatch();
  }
}
