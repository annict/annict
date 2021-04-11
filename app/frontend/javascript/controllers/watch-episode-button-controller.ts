import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  static classes = ['loading']
  static values = {
    recordId: Number
  }

  episodeId!: number | null;
  isLoading!: boolean;
  isWatched!: boolean;
  loadingClass!: string;
  pageCategory!: string | null;
  recordIdValue!: number | null;

  initialize() {
    this.episodeId = Number(this.data.get('episodeId'));
    this.pageCategory = this.data.get('pageCategory');
  }

  render() {
    if (this.isWatched) {
      this.element.classList.add('btn-info');
      this.element.classList.remove('btn-outline-info');
    } else {
      this.element.classList.remove('btn-info');
      this.element.classList.add('btn-outline-info');
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

    if (this.isWatched) {
      axios
        .delete(`/api/internal/records/${this.recordIdValue}`)
        .then(() => {
          this.endLoading()
          this.recordIdValue = null
          this.isWatched = false;
          this.render();
        });
    } else {
      axios
        .post('/api/internal/episode_records', {
          episode_id: this.episodeId,
          page_category: this.pageCategory,
        })
        .then((res: any) => {
          this.recordIdValue = res.data.record_id
          this.isWatched = true;
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
