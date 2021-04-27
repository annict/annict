import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  static classes = [
    'notReceivedButton',
    'receivedButton',
  ]
  static targets = ['iconWrapper'];
  static values = {
    channelId: Number,
    notReceivedIcon: String,
    receivedIcon: String,
  }

  channelIdValue!: number;
  currentReceivedValue!: boolean;
  iconWrapperTarget!: HTMLElement;
  notReceivedButtonClass!: string;
  receivedButtonClass!: string;
  receivedChannelIds!: number[]

  initialize() {
    this.element.classList.add('c-spinner');

    document.addEventListener('component-value-fetcher:receive-channel-button:fetched', (event: any) => {
      this.receivedChannelIds = event.detail;
      this.currentReceivedValue = this.receivedChannelIds.includes(this.channelIdValue)
      this.render()
    });
  }

  render() {
    if (this.currentReceivedValue) {
      this.element.classList.add(this.receivedButtonClass)
      this.element.classList.remove(this.notReceivedButtonClass)
      this.iconWrapperTarget.innerHTML = `<i class="fal fa-minus"></i>`
    } else {
      this.element.classList.remove(this.receivedButtonClass)
      this.element.classList.add(this.notReceivedButtonClass)
      this.iconWrapperTarget.innerHTML = '<i class="fal fa-plus"></i>'
    }

    this.element.classList.remove('c-spinner');
  }

  changeToReceived() {
    this.currentReceivedValue = true
  }

  changeToNotReceived() {
    this.currentReceivedValue = false
  }

  toggle() {
    this.element.setAttribute('disabled', 'true');

    if (this.currentReceivedValue) {
      axios
        .delete(`/api/internal/receptions/${this.channelIdValue}`)
        .then(() => {
          this.changeToNotReceived()
          this.element.removeAttribute('disabled');
          this.render()
        });
    } else {
      axios
        .post(`/api/internal/receptions`, {
          channel_id: this.channelIdValue,
        })
        .then(() => {
          this.changeToReceived()
          this.element.removeAttribute('disabled');
          this.render()
        });
    }
  }
}
