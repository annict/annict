import { Controller } from 'stimulus';

export default class extends Controller {
  static values = { eventName: String, url: String }

  boundReload!: any;
  eventNameValue!: string;
  urlValue!: string;

  initialize() {
    this.boundReload = this.reload.bind(this);
  }

  connect() {
    document.addEventListener(`reloadable--${this.eventNameValue}:reload`, this.boundReload);
  }

  disconnect() {
    document.removeEventListener(`reloadable--${this.eventNameValue}:reload`, this.boundReload);
  }

  async reload() {
    const html = await (await fetch(this.urlValue)).text();
    this.element.innerHTML = html
  }
}
