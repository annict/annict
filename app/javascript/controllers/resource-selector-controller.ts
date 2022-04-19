import { Controller } from '@hotwired/stimulus';
import Choices from 'choices.js';
import debounce from 'lodash/debounce';

export default class extends Controller {
  initialize() {
    let choice = new Choices(this.element as HTMLSelectElement, {
      searchResultLimit: 15,
      searchChoices: false,
    });

    this.element.addEventListener(
      'search',
      debounce((event: { detail: { value: any } }) => {
        if (event.detail.value) {
          choice.clearChoices();

          const url = `${this.requestPath()}?q=${event.detail.value}`;
          fetch(url)
            .then((res) => {
              return res.json();
            })
            .then((data) => {
              choice.setChoices(
                data.resources.map((r: { id: number; text: string }) => {
                  return { value: r.id, label: r.text };
                }),
                'value',
                'label',
                true,
              );
            });
        }
      }, 300),
    );
  }

  requestPath() {
    const modelName = this.data.get('modelName');

    if (!modelName) {
      return;
    }

    const paths: { [index: string]: string } = {
      Character: '/api/internal/characters',
      Organization: '/api/internal/organizations',
      Person: '/api/internal/people',
      Series: '/api/internal/series_list',
      Work: '/api/internal/works',
    };

    return paths[modelName];
  }
}
