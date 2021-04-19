import { Controller } from 'stimulus';

export default class extends Controller {
  static targets = ['errorPanel', 'errorMessageList'];

  errorPanelTarget!: HTMLElement;
  errorMessageListTarget!: HTMLElement;

  handleError(event: any) {
    const [errorMessages, _status, _xhr] = event.detail

    this.errorMessageListTarget.innerHTML = errorMessages.map((msg: string) => `<li>${msg}</li>`)
    this.errorPanelTarget.classList.remove('d-none')
  }
}
