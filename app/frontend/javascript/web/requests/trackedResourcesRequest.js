import eventHub from '../../common/eventHub';

export default {
  execute(_params) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/tracked_resources')
        .then((response) => {
          return response.json();
        })
        .then((trackedResources) => {
          eventHub.$emit('request:tracked-resources:fetched', trackedResources);
          resolve({ trackedResources });
        });
    });
  },
};
