import eventHub from '../../common/eventHub';

export default {
  execute(_params) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/library_entries')
        .then((response) => {
          return response.json();
        })
        .then((libraryEntries) => {
          eventHub.$emit('request:libraryEntries:fetched', libraryEntries);
          resolve({ libraryEntries });
        });
    });
  },
};
