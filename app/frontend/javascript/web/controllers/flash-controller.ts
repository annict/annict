import 'bootstrap/js/dist/alert';

import { Controller } from 'stimulus';

export default class extends Controller {
  type!: string | null;
  message!: string | null;

  initialize() {
    this.type = this.data.get('type');
    this.message = this.data.get('message');

    if (this.type && this.message) {
      this.displayMessage();
    }

    document.addEventListener('flash:show', (event: any) => {
      this.type = event.detail.type;
      this.message = event.detail.message;

      this.displayMessage();
    });
  }

  displayMessage() {
    const alertIconElm = this.element.querySelector('.c-flash__alert-icon');
    const messageElm = this.element.querySelector('.c-flash__message');

    if (alertIconElm && messageElm) {
      this.element.classList.remove('d-none');
      this.element.classList.add(this.alertClass);
      alertIconElm.classList.add(this.alertIconClass);
      messageElm.innerHTML = this.message || '';
    }
  }

  get alertClass() {
    switch (this.type) {
      case 'notice':
        return 'alert-success';
      case 'alert':
        return 'alert-danger';
      default:
        return '';
    }
  }

  get alertIconClass() {
    switch (this.type) {
      case 'notice':
        return 'fa-check-circle';
      case 'alert':
        return 'fa-exclamation-triangle';
      default:
        return '';
    }
  }
}
