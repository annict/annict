import Choices from 'choices.js';
import debounce from 'lodash.debounce';

export default {
  template: `
    <select :name="name">
      <slot></slot>
    </select>
  `,

  props: {
    name: {
      type: String,
      required: true,
    },
    modelName: {
      type: String,
      required: true,
    },
  },

  computed: {
    requestPath() {
      return {
        Character: '/api/internal/characters',
        Organization: '/api/internal/organizations',
        Person: '/api/internal/people',
        Series: '/api/internal/series_list',
        Work: '/api/internal/works',
      }[this.modelName];
    },
  },

  mounted() {
    let choice = new Choices(this.$el, {
      searchResultLimit: 15,
      searchChoices: false,
    });

    this.$el.addEventListener(
      'search',
      debounce(event => {
        if (event.detail.value) {
          choice.clearChoices();

          const url = `${this.requestPath}?q=${event.detail.value}`;
          fetch(url)
            .then(res => {
              return res.json();
            })
            .then(data => {
              choice.setChoices(
                data.resources.map(r => {
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
  },
};
