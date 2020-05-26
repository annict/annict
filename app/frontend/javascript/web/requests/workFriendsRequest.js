import eventHub from '../../common/eventHub';
import urlParams from "../utils/urlParams";

export default {
  execute(params) {
    return new Promise((resolve, reject) => {
      const query = urlParams({
        work_ids: params.work_ids,
        display_option: params.display_option
      })

      fetch(`/api/internal/work_friends?${query}`)
        .then((response) => {
          return response.json();
        })
        .then((result) => {
          eventHub.$emit('request:work-friends:fetched', result);
          resolve({ result });
        });
    });
  },
};
