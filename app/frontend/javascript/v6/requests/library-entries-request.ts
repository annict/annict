import { EventDispatcher } from '../utils/event-dispatcher';

export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/library_entries')
        .then((response) => {
          return response.json();
        })
        .then((libraryEntries) => {
          new EventDispatcher('user-data-fetcher:library-entries:fetched', libraryEntries).dispatch();
          resolve({ libraryEntries });
        });
    });
  },
};
