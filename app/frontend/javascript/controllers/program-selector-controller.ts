import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static targets = ['select'];
  static values = { libraryEntryId: Number, initProgramId: String }

  libraryEntryIdValue!: number;
  initProgramIdValue!: string;
  selectTarget!: HTMLSelectElement;
  currentProgramId!: string;

  initialize() {
    if (this.initProgramIdValue !== '') {
      this.selectTarget.value = this.initProgramIdValue
      this.currentProgramId = this.initProgramIdValue
    } else {
      this.selectTarget.value = NO_SELECT
    }
  }

  reloadList() {
    new EventDispatcher('reloadable--trackable-episode-list:reload').dispatch();
  }

  change() {
    const newProgramId = this.selectTarget.value

    if (newProgramId !== this.currentProgramId) {
      this.element.setAttribute('disabled', 'true');

      axios
        .patch(`/api/internal/library_entries/${this.libraryEntryIdValue}`, {
          program_id: newProgramId,
        })
        .then(() => {
          this.currentProgramId = newProgramId
          this.element.removeAttribute('disabled');
          this.reloadList()
        });
    }
  }
}
