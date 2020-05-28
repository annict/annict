export default {
  execute(_params: any) {
    return new Promise((resolve, reject) => {
      fetch('/api/internal/following')
        .then((response) => {
          return response.json();
        })
        .then((following) => {
          // eventHub.$emit('request:following:fetched', following);
          resolve({ following });
        });
    });
  },
};
