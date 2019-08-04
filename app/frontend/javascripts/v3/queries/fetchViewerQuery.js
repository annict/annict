import gql from 'graphql-tag'

import client from '../client'

const query = gql`
  query fetchViewer {
    viewer {
      username
      avatarUrl(size: size50)
      locale
    }
  }
`

export default () => {
  return client.query({ query: query })
}
