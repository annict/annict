import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['input'];

  inputTarget!: HTMLElement;

  initialize() {
    this.resetState();
  }

  changeState(event: Event) {
    const { state } = event.currentTarget.dataset;

    this.resetState();

    if (this.inputTarget.value === state) {
      this.inputTarget.value = null;
    } else {
      event.currentTarget.classList.remove('u-btn-outline-input-border');
      event.currentTarget.classList.add(`u-btn-${state}`);

      this.inputTarget.value = state;
    }
  }

  resetState() {
    const buttonElms = this.element.getElementsByClassName('btn');

    for (let buttonElm of buttonElms) {
      const { state } = buttonElm.dataset;

      buttonElm.classList.remove(`u-btn-${state}`);
      buttonElm.classList.add('u-btn-outline-input-border');
    }
  }
}
