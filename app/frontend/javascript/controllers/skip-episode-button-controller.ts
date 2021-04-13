import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static classes = ['loading']

  episodeId!: number | null;
  isLoading!: boolean;
  isSkipped!: boolean;
  loadingClass!: string;

  initialize() {
    this.episodeId = Number(this.data.get('episodeId'));
  }

  render() {
    if (this.isSkipped) {
      this.element.classList.add('btn-secondary');
      this.element.classList.remove('btn-outline-secondary');
    } else {
      this.element.classList.remove('btn-secondary');
      this.element.classList.add('btn-outline-secondary');
    }
  }

  startLoading() {
    this.isLoading = true
    this.element.classList.add(this.loadingClass);
  }

  endLoading() {
    this.isLoading = false
    this.element.classList.remove(this.loadingClass);
  }

  toggle() {
    if (this.isLoading) {
      return;
    }

    this.startLoading()

    if (this.isSkipped) {
      axios
        .delete(`/api/internal/skipped_episodes/${this.episodeId}`)
        .then(() => {
          this.endLoading()
          this.isSkipped = false;

          new EventDispatcher('tracking-state:change-to-untracked', {
            episodeId: this.episodeId
          }).dispatch();

          this.render();
        });
    } else {
      axios
        .post('/api/internal/skipped_episodes', {
          episode_id: this.episodeId,
        })
        .then(() => {
          this.isSkipped = true;

          new EventDispatcher('tracking-state:change-to-tracked', {
            episodeId: this.episodeId
          }).dispatch();
        })
        .catch(() => {
          ($('.c-sign-up-modal') as any).modal('show');
        })
        .then(() => {
          this.endLoading()
          this.render();
        });
    }
  }
}
