import 'bootstrap/js/dist/alert';

import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  type!: string;
  message!: string;

  initialize() {
    this.type = this.data.get('type') || 'notice';
    this.message = this.data.get('message') || '';

    if (this.type && this.message) {
      this.displayMessage();
    }

    document.addEventListener('flash:show', ({ detail: { type, message } }: any) => {
      this.type = type || 'notice';
      this.message = message || '';

      this.displayMessage();
    });
  }

  displayMessage() {
    const alertIconElm = this.element.querySelector('.c-flash__alert-icon');
    const messageElm = this.element.querySelector('.c-flash__message');

    if (alertIconElm && messageElm) {
      this.element.classList.remove('d-none');
      this.element.classList.add(this.alertClass);
      alertIconElm.innerHTML = this.alertIcon; // Using innerHTML to render fontawesome icon after refetch
      messageElm.innerHTML = this.message;
    }
  }

  close() {
    this.element.classList.add('d-none');
  }

  get alertClass() {
    switch (this.type) {
      case 'alert':
        return 'alert-danger';
      default:
        return 'alert-success';
    }
  }

  get alertIcon() {
    switch (this.type) {
      case 'alert':
        return '<i class="far fa-exclamation-triangle"></i>';
      default:
        return '<i class="far fa-check-circle"></i>';
    }
  }
}
