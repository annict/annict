import gql from 'graphql-tag'

import client from '../client'
import { Channel } from '../models'
import { ApplicationQuery } from './ApplicationQuery'

const query = gql`
  query {
    channels(isVod: true) {
      nodes {
        annictId
        name
      }
    }
  }
`

export class FetchVodChannelsQuery extends ApplicationQuery {
  public async execute(): Promise<[Channel]> {
    const result = await client.query({ query: query })
    return result.data.channels.nodes.map((node): Channel => {
      return new Channel(node)
    })
  }
}
