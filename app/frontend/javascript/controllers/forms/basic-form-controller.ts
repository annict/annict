import { Controller } from 'stimulus';

import { EventDispatcher } from '../../utils/event-dispatcher';

export default class extends Controller {
  static targets = ['errorPanel', 'errorMessageList', 'form', 'submitButton'];

  errorPanelTarget!: HTMLElement;
  errorMessageListTarget!: HTMLElement;
  formTarget!: HTMLFormElement;
  submitButtonTarget!: HTMLElement;

  handleSubmitStart(_event: any) {
    this.submitButtonTarget.setAttribute('disabled', 'true');
    this.submitButtonTarget.classList.add('c-spinner');
  }

  async handleSubmitEnd(event: any) {
    this.submitButtonTarget.removeAttribute('disabled');
    this.submitButtonTarget.classList.remove('c-spinner');

    const { success } = event.detail

    if (success) {
      await this.handleSuccess(event)
    } else {
      await this.handleError(event)
    }
  }

  async handleSuccess(event: any) {
    this.formTarget.reset()

    const { fetchResponse } = event.detail
    const data = JSON.parse(await fetchResponse.responseText)

    if (data && data.flash) {
      new EventDispatcher('flash:show', data.flash).dispatch();
    }
  };

  async handleError(event: any) {
    const { fetchResponse } = event.detail
    const errorMessages = JSON.parse(await fetchResponse.responseText)

    this.errorMessageListTarget.innerHTML = errorMessages.map((msg: string) => `<li>${msg}</li>`)
    this.errorPanelTarget.classList.remove('d-none')
  }
}
