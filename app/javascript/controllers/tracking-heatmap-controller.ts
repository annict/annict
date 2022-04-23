import Tooltip from 'bootstrap/js/dist/tooltip';
import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static values = {
    framePath: String,
  };

  framePathValue!: string;

  async connect() {
    this.element.setAttribute('src', this.framePathValue);

    await (this.element as any).loaded;

    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map((tooltipTriggerEl: any) => {
      return new Tooltip(tooltipTriggerEl);
    });
  }
}
