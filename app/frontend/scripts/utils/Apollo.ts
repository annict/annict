import { InMemoryCache } from 'apollo-cache-inmemory'
import { ApolloClient } from 'apollo-client'
import { HttpLink } from 'apollo-link-http'
import VueApollo from 'vue-apollo'

import './Global'

export default {
  client() {
    const httpLink = new HttpLink({
      uri: `${window.ann.BASE_DATA.API_URL}/graphql`,
    })

    return new ApolloClient({
      cache: new InMemoryCache(),
      connectToDevTools: true,
      link: httpLink,
    })
  },

  provider() {
    return new VueApollo({
      defaultClient: this.client(),
    })
  },
}
