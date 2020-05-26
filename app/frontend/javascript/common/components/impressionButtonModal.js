import $ from 'jquery';
import uniq from 'lodash/uniq';

import eventHub from '../../common/eventHub';

export default {
  template: '#t-impression-button-modal',

  data() {
    return {
      workId: null,
      tagNames: [],
      allTagNames: [],
      popularTagNames: [],
      comment: '',
      isLoading: false,
      isSaving: false,
    };
  },

  methods: {
    load() {
      this.isLoading = true;

      return $.ajax({
        method: 'GET',
        url: '/api/internal/impression',
        data: {
          work_id: this.workId,
        },
      })
        .done((data) => {
          this.tagNames = data.tag_names;
          this.allTagNames = data.all_tag_names;
          this.popularTagNames = data.popular_tag_names;
          this.comment = data.comment;

          return setTimeout(() => {
            const $tagsInput = $('.js-impression-tags');
            $tagsInput.select2({
              tags: true,
            });
            $tagsInput.on('select2:select', (event) => {
              return (this.tagNames = $(event.currentTarget).val());
            });
            return $tagsInput.on('select2:unselect', (event) => {
              return (this.tagNames = $(event.currentTarget).val() || []);
            });
          });
        })
        .fail(function () {
          const message = gon.I18n['messages._components.impression_button.error'];
          return eventHub.$emit('flash:show', message, 'alert');
        })
        .always(() => {
          return (this.isLoading = false);
        });
    },

    add(tagName) {
      const $tagsInput = $('.js-impression-tags');

      this.allTagNames.push(tagName);
      this.allTagNames = uniq(this.allTagNames);
      this.tagNames.push(tagName);
      this.tagNames = uniq(this.tagNames);

      $tagsInput.val(this.tagNames);
      return $tagsInput.trigger('change');
    },

    save() {
      this.isSaving = true;

      return $.ajax({
        method: 'PATCH',
        url: '/api/internal/impression',
        data: {
          work_id: this.workId,
          tags: this.tagNames,
          comment: this.comment,
        },
      })
        .done((data) => {
          $('.c-impression-button-modal').modal('hide');
          eventHub.$emit('workTags:saved', this.workId, data.tags);
          eventHub.$emit('workComment:saved', this.workId, data.comment);
          return eventHub.$emit('flash:show', gon.I18n['messages._common.updated']);
        })
        .fail(function () {
          const message = gon.I18n['messages._components.impression_button.error'];
          return eventHub.$emit('flash:show', message, 'alert');
        })
        .always(() => {
          return (this.isSaving = false);
        });
    },
  },

  created() {
    return eventHub.$on('impressionButtonModal:show', (workId) => {
      this.workId = workId;
      this.load();
      return $('.c-impression-button-modal').modal('show');
    });
  },
};
