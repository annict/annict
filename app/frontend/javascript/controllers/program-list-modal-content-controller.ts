import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static targets = ['programId'];

  programIdTargets!: [HTMLInputElement];

  check() {
    const checkedProgramId = this.programIdTargets.find(target => target.checked)?.value

    if (!checkedProgramId) {
      new EventDispatcher('program-list-modal:close').dispatch();
      return
    }

    axios
      .post('/api/internal/program_checks', {
        program_id: checkedProgramId,
      })
      .then((res) => {
        new EventDispatcher('program-list-modal:close').dispatch();
      });
  }

  uncheck(event: any) {
    const animeId = this.data.get('animeId')

    if (!animeId) {
      new EventDispatcher('program-list-modal:close').dispatch();
      return
    }

    event.target.classList.add('c-spinner');

    axios
      .delete('/api/internal/program_checks', {
        params: {
          anime_id: animeId
        },
      })
      .then((res) => {
        new EventDispatcher('program-list-modal:close').dispatch();
      });
  }
}
