import Modal from 'bootstrap/js/dist/modal';
import { Controller } from '@hotwired/stimulus';

import fetcher from '../utils/fetcher';

export default class extends Controller {
  static classes = ['notReceivedButton', 'receivedButton'];
  static targets = ['iconWrapper'];
  static values = {
    channelId: Number,
    notReceivedIcon: String,
    receivedIcon: String,
  };

  channelIdValue!: number;
  currentReceivedValue!: boolean;
  iconWrapperTarget!: HTMLElement;
  notReceivedButtonClass!: string;
  receivedButtonClass!: string;
  receivedChannelIds!: number[];

  initialize() {
    this.element.classList.add('c-spinner');

    document.addEventListener('component-value-fetcher:receive-channel-button:fetched', (event: any) => {
      this.receivedChannelIds = event.detail;
      this.currentReceivedValue = this.receivedChannelIds.includes(this.channelIdValue);
      this.render();
    });
  }

  render() {
    if (this.currentReceivedValue) {
      this.element.classList.add(this.receivedButtonClass);
      this.element.classList.remove(this.notReceivedButtonClass);
      this.iconWrapperTarget.innerHTML = `<i class="fal fa-minus"></i>`;
    } else {
      this.element.classList.remove(this.receivedButtonClass);
      this.element.classList.add(this.notReceivedButtonClass);
      this.iconWrapperTarget.innerHTML = '<i class="fal fa-plus"></i>';
    }

    this.element.classList.remove('c-spinner');
    this.element.removeAttribute('disabled');
  }

  changeToReceived() {
    this.currentReceivedValue = true;
  }

  changeToNotReceived() {
    this.currentReceivedValue = false;
  }

  async toggle() {
    this.element.setAttribute('disabled', 'true');

    try {
      if (this.currentReceivedValue) {
        const data = await fetcher.delete(`/api/internal/channels/${this.channelIdValue}/reception`);

        this.changeToNotReceived();
      } else {
        const data = await fetcher.post(`/api/internal/channels/${this.channelIdValue}/reception`);

        this.changeToReceived();
      }
    } catch (err) {
      if (err.response.status === 401) {
        new Modal('.c-sign-up-modal').show();
      }
    } finally {
      this.render();
    }
  }
}
