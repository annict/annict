import eventHub from '../../common/eventHub';

export default {
  execute() {
    new Promise((resolve, reject) => {
      fetch('/api/internal/viewer')
        .then((response) => {
          return response.json();
        })
        .then((viewer) => {
          eventHub.$emit('request:viewer:fetched', viewer);
          resolve();
        });
    });
  },
};
