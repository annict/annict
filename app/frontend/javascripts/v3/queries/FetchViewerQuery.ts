import gql from 'graphql-tag'

import client from '../client'
import { ApplicationQuery } from './ApplicationQuery'

const query = gql`
  query fetchViewer {
    viewer {
      username
      avatarUrl(size: size50)
      locale
      isCommitter
    }
  }
`

export class FetchViewerQuery extends ApplicationQuery {
  public async execute() {
    return client.query({ query })
  }
}
