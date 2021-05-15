import { EventDispatcher } from '../utils/event-dispatcher';

export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/following')
        .then((response) => {
          return response.json();
        })
        .then((following) => {
          new EventDispatcher('user-data-fetcher:following:fetched', { following }).dispatch();
          resolve({ following });
        });
    });
  },
};
