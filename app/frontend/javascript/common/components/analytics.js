import analytics from '../analytics';

export default function(event) {
  return {
    template: '<div></div>',

    props: {
      trackingId: {
        type: String,
        required: true,
      },
      clientId: {
        type: String,
        required: true,
      },
      userId: {
        type: String,
        required: true,
      },
      dimension1: {
        type: String,
        required: true,
      },
      dimension2: {
        type: String,
        required: true,
      },
    },

    methods: {
      create() {
        const options = {
          storage: 'none',
          clientId: this.clientId,
        };

        if (this.userId) {
          options['userId'] = this.userId;
        }

        if (typeof ga === 'function') {
          return ga('create', this.trackingId, options);
        }
      },

      send() {
        if (typeof ga === 'function') {
          ga('set', 'location', event.data.url);
          return ga('send', 'pageview', {
            dimension1: this.dimension1,
            dimension2: this.dimension2,
          });
        }
      },
    },

    mounted() {
      analytics.load();
      this.create();
      return this.send();
    },
  };
}
