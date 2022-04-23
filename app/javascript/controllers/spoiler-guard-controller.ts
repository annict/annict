import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static values = { workId: Number, episodeId: Number };

  episodeIdValue!: number;
  isNotSpoiler!: boolean;
  workIdValue!: number;

  initialize() {
    document.addEventListener('component-value-fetcher:spoiler-guard:fetched', (event: any) => {
      const {
        is_signed_in,
        hide_record_body,
        watched_work_ids,
        work_ids_in_library,
        tracked_episode_ids,
      } = event.detail;

      this.isNotSpoiler =
        !is_signed_in ||
        !hide_record_body ||
        !work_ids_in_library.includes(this.workIdValue) ||
        (!this.episodeIdValue && watched_work_ids.includes(this.workIdValue)) ||
        (this.episodeIdValue && tracked_episode_ids.includes(this.episodeIdValue));

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
