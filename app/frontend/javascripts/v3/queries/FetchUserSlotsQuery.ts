import gql from 'graphql-tag'

import client from '../client'
import { Slot } from '../models'
import { ApplicationQuery } from './ApplicationQuery'

const query = gql`
  query {
    viewer {
      slots(watched: false, orderBy: { field: STARTED_AT, direction: DESC }, first: 20) {
        nodes {
          startedAt
          rebroadcast
          channel {
            name
          }
          work {
            annictId
            localTitle
            image {
              internalUrl(size: "48x")
            }
          }
          episode {
            annictId
            numberText
            title
          }
        }
      }
    }
  }
`

export class FetchUserSlotsQuery extends ApplicationQuery {
  public async execute(): Promise<[Slot]> {
    const result = await client.query({ query })

    return result.data.viewer.slots.nodes.map(node => {
      return new Slot(node)
    })
  }
}
