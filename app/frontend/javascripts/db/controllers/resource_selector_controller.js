import Choices from 'choices.js';
import { Controller } from 'stimulus';
import debounce from 'lodash.debounce';

export default class extends Controller {
  static targets = ['selectBox'];

  initialize() {
    this.isLoading = false
  }

  connect() {
    let choice = new Choices(this.selectBoxTarget, {
      searchResultLimit: 15,
      searchChoices: false
    });

    this.selectBoxTarget.addEventListener('search', debounce((event) => {
      if (event.detail.value) {
        choice.clearChoices();

        if (this.isLoading) {
          return;
        }
        this.isLoading = true;

        const url = `${this.requestPath(this.data.get('modelName'))}?q=${event.detail.value}`;
        fetch(url)
          .then(res => {
            return res.json();
          })
          .then(data => {
            choice.setChoices(
              data.resources.map(r => {
                return { value: r.id, label: r.text }
              }),
              'value',
              'label',
              true,
            );

            this.isLoading = false;
          });
      }
    }, 300));
  }

  requestPath(modelName) {
    return {
      Character: '/api/internal/characters',
      Organization: '/api/internal/organizations',
      Person: '/api/internal/people',
      Series: '/api/internal/series_list',
      Work: '/api/internal/works',
    }[modelName];
  }
}
