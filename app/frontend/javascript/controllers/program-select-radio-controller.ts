import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static values = { libraryEntryId: Number, initProgramId: Number };

  libraryEntryIdValue!: number;
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

  change(event: any) {
    const newProgramId = event.currentTarget.value;

    if (newProgramId !== this.currentProgramId) {
      this.toggleLoading(true);

      axios
        .patch(`/api/internal/library_entries/${this.libraryEntryIdValue}`, {
          program_id: newProgramId,
        })
        .then(() => {
          this.currentProgramId = newProgramId;
          this.toggleLoading(false);
          this.reloadList();
        });
    }
  }
}
