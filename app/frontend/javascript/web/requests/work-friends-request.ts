import urlParams from "../utils/url-params";

export default {
  execute(params: { work_ids: number[], display_option: string }) {
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
          // eventHub.$emit('request:work-friends:fetched', result);
          resolve({ result });
        });
    });
  },
};
