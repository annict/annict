import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  episodeId!: string;
  isLoading!: boolean;

  initialize() {
    this.episodeId = this.data.get('episodeId') ?? '';
    this.isLoading = false;
  }

  track() {
    if (this.isLoading) {
      return;
    }

    this.isLoading = true;
    this.showSpinner();

    axios
      .post(`/api/internal/episode_records`, {
        episode_id: this.episodeId,
      })
      .then(() => {
        this.isLoading = false;
        this.hideSpinner();
      });
  }

  showSpinner() {
    this.element.setAttribute('disabled', '');
    this.element.classList.add('c-spinner');
  }

  hideSpinner() {
    this.element.removeAttribute('disabled');
    this.element.classList.remove('c-spinner');
  }
}
