import { EventDispatcher } from '../utils/event-dispatcher';

export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/tracked_resources')
        .then((response) => {
          return response.json();
        })
        .then((trackedResources) => {
          new EventDispatcher('user-data-fetcher:tracked-resources:fetched', { trackedResources }).dispatch();
          resolve({ trackedResources });
        });
    });
  },
};
