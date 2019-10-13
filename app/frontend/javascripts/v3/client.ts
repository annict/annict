import { ApolloClient } from 'apollo-client'
import { InMemoryCache, IntrospectionFragmentMatcher } from 'apollo-cache-inmemory'
import { HttpLink } from 'apollo-link-http'

import introspectionQueryResultData from './fragmentTypes.json'

const fragmentMatcher = new IntrospectionFragmentMatcher({
  introspectionQueryResultData,
})

const cache = new InMemoryCache({ fragmentMatcher })
const csrfTokenElm = document.querySelector('meta[name=csrf-token]')

let headers = {}
if (csrfTokenElm) {
  headers = {
    'X-CSRF-Token': csrfTokenElm.getAttribute('content'),
  }
}

const link = new HttpLink({
  uri: '/api/internal/graphql',
  headers
})

export default new ApolloClient({
  cache,
  link,
})
