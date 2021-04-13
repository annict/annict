import { Controller } from 'stimulus';

export default class extends Controller {
  static classes = ['tracked']
  static values = {
    episodeId: Number
  }

  boundChangeToTracked!: any;
  boundChangeToUntracked!: any;
  episodeIdValue!: number | null;
  trackedClass!: string;

  initialize() {
    this.boundChangeToTracked = this.changeToTracked.bind(this);
    this.boundChangeToUntracked = this.changeToUntracked.bind(this);
  }

  connect() {
    document.addEventListener('tracking-state:change-to-tracked', this.boundChangeToTracked);
    document.addEventListener('tracking-state:change-to-untracked', this.boundChangeToUntracked);
  }

  disconnect() {
    document.removeEventListener('tracking-state:change-to-tracked', this.boundChangeToTracked);
    document.removeEventListener('tracking-state:change-to-untracked', this.boundChangeToUntracked);
  }

  changeToTracked({ detail: { episodeId } }: any) {
    if (this.episodeIdValue === episodeId) {
      this.element.classList.add(this.trackedClass)
    }
  }

  changeToUntracked({ detail: { episodeId } }: any) {
    if (this.episodeIdValue === episodeId) {
      this.element.classList.remove(this.trackedClass)
    }
  }
}
