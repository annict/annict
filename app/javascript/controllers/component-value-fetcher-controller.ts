import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import fetcher from '../utils/fetcher';

export default class extends Controller {
  static values = {
    payload: Object,
    eventName: String,
    method: String,
    url: String,
  };

  payloadValue!: object;
  eventNameValue!: string;
  methodValue!: string;
  urlValue!: string;

  async connect() {
    let data;

    if (this.methodValue === 'get') {
      data = await fetcher.get(this.urlValue);
    } else if (this.methodValue === 'post') {
      data = await fetcher.post(this.urlValue, this.payloadValue);
    }

    new EventDispatcher(this.eventNameValue, data).dispatch();
  }
}
