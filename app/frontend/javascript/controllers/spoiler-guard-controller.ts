import { Controller } from 'stimulus';

export default class extends Controller {
  workId!: number;
  episodeId!: number;
  libraryEntries!: { work_id: number; status_kind: string }[];
  trackedResources!: any;
  isSpoiler!: boolean;

  initialize() {
    this.workId = Number(this.data.get('workId'));
    this.episodeId = Number(this.data.get('episodeId'));
    this.isSpoiler = true;

    document.addEventListener(
      'user-data-fetcher:fetched-all',
      ({ detail: { libraryEntries, trackedResources } }: any) => {
        this.libraryEntries = libraryEntries;
        this.trackedResources = trackedResources;

        if (!this.libraryEntries || !this.trackedResources) {
          return;
        }

        this.checkSpoiler();
        this.render();
      },
    );
  }

  checkSpoiler() {
    const work = this.libraryEntries.filter((entry) => {
      return entry.work_id === this.workId;
    })[0];

    if (!work || work.status_kind === 'watched' || work.status_kind === 'stop_watching') {
      this.isSpoiler = false;
      return;
    }

    const trackedWorkIds = this.trackedResources.work_ids;
    if (trackedWorkIds.includes(this.workId)) {
      this.isSpoiler = false;
      return;
    }

    const trackedEpisodeIds = this.trackedResources.episode_ids;
    if (this.episodeId && trackedEpisodeIds.includes(this.episodeId)) {
      this.isSpoiler = false;
    }
  }

  render() {
    if (this.isSpoiler) {
      this.element.classList.add('is-spoiler');
      this.element.classList.remove('is-not-spoiler');
    } else {
      this.element.classList.remove('is-spoiler');
      this.element.classList.add('is-not-spoiler');
    }
  }

  hide() {
    this.isSpoiler = false;
    this.render();
  }
}
