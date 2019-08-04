import { ApolloClient } from 'apollo-client'
import { InMemoryCache } from 'apollo-cache-inmemory'
import { HttpLink } from 'apollo-link-http'

const cache = new InMemoryCache()
const link = new HttpLink({
  uri: 'http://annict-jp.test:3000/api/internal/graphql',
})

export default new ApolloClient({
  cache,
  link,
})
