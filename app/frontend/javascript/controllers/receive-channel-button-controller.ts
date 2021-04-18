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
    initReceived: Boolean,
    notReceivedIcon: String,
    receivedIcon: String,
  }

  channelIdValue!: number;
  currentReceivedValue!: boolean;
  iconWrapperTarget!: HTMLElement;
  initReceivedValue!: boolean;
  notReceivedButtonClass!: string;
  receivedButtonClass!: string;

  initialize() {
    this.currentReceivedValue = this.initReceivedValue
  }

  changeToReceived() {
    this.currentReceivedValue = true
    this.element.classList.add(this.receivedButtonClass)
    this.element.classList.remove(this.notReceivedButtonClass)
    this.iconWrapperTarget.innerHTML = `<i class="fal fa-minus"></i>`
  }

  changeToNotReceived() {
    this.currentReceivedValue = false
    this.element.classList.remove(this.receivedButtonClass)
    this.element.classList.add(this.notReceivedButtonClass)
    this.iconWrapperTarget.innerHTML = '<i class="fal fa-plus"></i>'
  }

  toggle() {
    this.element.setAttribute('disabled', 'true');

    if (this.currentReceivedValue) {
      axios
        .delete(`/api/internal/receptions/${this.channelIdValue}`)
        .then(() => {
          this.changeToNotReceived()
          this.element.removeAttribute('disabled');
        });
    } else {
      axios
        .post(`/api/internal/receptions`, {
          channel_id: this.channelIdValue,
        })
        .then(() => {
          this.changeToReceived()
          this.element.removeAttribute('disabled');
        });
    }
  }
}
