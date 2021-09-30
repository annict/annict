import axios from 'axios';
import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import fetcher from '../utils/fetcher';

export default class extends Controller {
  static values = { workId: Number, initProgramId: Number };

  workIdValue!: number;
  initProgramIdValue!: number;
  currentProgramId!: number;

  initialize() {
    this.currentProgramId = this.initProgramIdValue;
  }

  reloadList() {
    new EventDispatcher('reloadable--trackable-episode-list:reload').dispatch();
  }

  toggleLoading(disabled: boolean) {
    this.element.querySelectorAll('.form-check-input').forEach((radioElm) => {
      (radioElm as HTMLInputElement).disabled = disabled;
    });
  }

  async change(event: any) {
    const newProgramId = event.currentTarget.value;

    if (newProgramId !== this.currentProgramId) {
      this.toggleLoading(true);

      try {
        await fetcher.post(`/api/internal/works/${this.workIdValue}/program_select`, {
          program_id: newProgramId,
        });

        this.currentProgramId = newProgramId;
        this.reloadList();
      } catch (err) {
        console.error(err);
      } finally {
        this.toggleLoading(false);
      }
    }
  }
}
