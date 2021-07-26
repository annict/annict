import { Controller } from 'stimulus';

export default class extends Controller {
  static values = { workId: Number, episodeId: Number };

  episodeIdValue!: number;
  isNotSpoiler!: boolean;
  workIdValue!: number;

  initialize() {
    document.addEventListener('component-value-fetcher:spoiler-guard:fetched', (event: any) => {
      const { is_signed_in, episode_ids, work_ids } = event.detail;

      this.isNotSpoiler =
        !is_signed_in ||
        (!this.episodeIdValue && work_ids.includes(this.workIdValue)) ||
        episode_ids.includes(this.episodeIdValue);

      this.render();
    });
  }

  render() {
    if (this.isNotSpoiler) {
      this.element.classList.remove('is-spoiler');
    }
  }

  hide() {
    this.isNotSpoiler = true;
    this.render();
  }
}
