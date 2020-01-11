import gql from 'graphql-tag'

import client from '../client'
import { LibraryEntry } from '../models'
import { ApplicationQuery } from './ApplicationQuery'

const query = gql`
  query {
    viewer {
      libraryEntries(filterBy: { statusKinds: [WATCHING] }, orderBy: { field: UPDATED_AT, direction: DESC }) {
        nodes {
          work {
            annictId
            title
          }
          untappedEpisodes {
            totalCount
          }
        }
      }
    }
  }
`

export class FetchLibraryEntriesQuery extends ApplicationQuery {
  public async execute(): Promise<[LibraryEntry]> {
    const result = await client.query({ query })

    return result.data.viewer.libraryEntries.nodes.map(node => {
      const libraryEntry = new LibraryEntry(node)
      libraryEntry.setWork(node.work)
      return libraryEntry
    })
  }
}
