import { Controller } from 'stimulus';

import { EventDispatcher } from '../../utils/event-dispatcher';

export default abstract class extends Controller {
  static values = {
    eventName: String,
    url: String,
  }

  eventNameValue!: string;
  urlValue!: string;

  connect() {
    fetch(this.urlValue)
    .then((res) => {
      return res.json();
    })
    .then((data) => {
      new EventDispatcher(this.eventNameValue, data).dispatch();
    });
  }
}
