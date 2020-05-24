import eventHub from '../../common/eventHub';

export default {
  execute() {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/likes')
        .then((response) => {
          return response.json();
        })
        .then((likes) => {
          eventHub.$emit('request:likes:fetched', likes);
          resolve({ likes });
        });
    });
  },
};
