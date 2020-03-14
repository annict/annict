import $ from 'jquery';
import _ from 'lodash';

export default {
  inserted(el, binding) {
    return $(el).select2({
      ajax: {
        url: _requestUrl(binding.value.model),
        delay: 250,
        data(params) {
          return { q: params.term };
        },
        processResults(data) {
          return {
            results: _.map(data.resources, resource => ({
              id: resource.id,
              text: resource.text,
            })),
          };
        },
        minimumInputLength: 1,
      },
    });
  },
};

var _requestUrl = function(model) {
  const urls = {
    Character: '/api/internal/characters',
    Organization: '/api/internal/organizations',
    Person: '/api/internal/people',
    Series: '/api/internal/series_list',
    Work: '/api/internal/works',
  };

  return urls[model];
};
