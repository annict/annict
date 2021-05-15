import { EventDispatcher } from '../utils/event-dispatcher';

export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/user')
        .then((response) => {
          return response.json();
        })
        .then((user) => {
          new EventDispatcher('user-data-fetcher:user:fetched', { user }).dispatch();
          resolve({ user });
        });
    });
  },
};
