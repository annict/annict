import { Controller } from '@hotwired/stimulus';

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

  targetUrl() {
    if (this.element.tagName === 'TURBO-FRAME' && this.element.hasAttribute('src')) {
      return this.element.getAttribute('src') ?? '/500.html'
    } else {
      return this.urlValue
    }
  }

  async reload() {
    const html = await (await fetch(this.targetUrl())).text();
    this.element.innerHTML = html
  }
}
