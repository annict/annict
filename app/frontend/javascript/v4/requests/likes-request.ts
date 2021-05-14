import { EventDispatcher } from '../utils/event-dispatcher';

export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/likes')
        .then((response) => {
          return response.json();
        })
        .then((likes) => {
          new EventDispatcher('user-data-fetcher:likes:fetched', { likes }).dispatch();
          resolve({ likes });
        });
    });
  },
};
