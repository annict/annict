import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  episodeId!: string | null;
  isTracked!: boolean;
  isLoading!: boolean;

  initialize() {
    this.episodeId = this.data.get('episodeId');
    this.isLoading = false
    this.isTracked = false
  }

  track() {
    if (this.isLoading && this.isTracked) {
      return;
    }

    this.isLoading = true;
    this.showSpinner()
    this.element.setAttribute('disabled', '')

    axios
      .post(`/api/internal/episode_records`, {
        episode_id: this.episodeId
      })
      .then(() => {
        this.isLoading = false;
        this.hideSpinner()
        this.makeBtnComplete()
        this.isTracked = true
      });
  }

  showSpinner() {
    this.element.classList.add('c-spinner');
  }

  hideSpinner() {
    this.element.classList.remove('c-spinner');
  }

  makeBtnComplete() {
    this.element.classList.remove('btn-outline-primary');
    this.element.classList.add('btn-primary');
  }
}
