import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

export default class extends Controller {
  initialize() {
    document.addEventListener('program-list-modal:open', ({ detail: { animeId } }: any) => {
      $(this.element).modal('show')

      axios
        .get('/api/internal/program_list_modal_content', {
          params: {
            anime_id: animeId,
          },
        })
        .then((res) => {
          const modalBodyElm = this.element.querySelector('.modal-body');

          if (modalBodyElm) {
            modalBodyElm.innerHTML = res.data;
          }
        });
    });

    document.addEventListener('program-list-modal:close', () => {
      $(this.element).modal('hide')
    });

    $(this.element).on('hidden.bs.modal', () => {
      $('.modal-backdrop').remove();
    });
  }
}
