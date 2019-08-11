import _ from 'lodash'
import gql from 'graphql-tag'

import client from '../client'
import Work from '../models/Work'

const query = gql`
  query($annictId: Int!) {
    works(annictIds: [$annictId]) {
      nodes {
        id
        annictId
        title
        titleKana
        titleEn
        media
        seasonName
        seasonSlug
        seasonYear
        startedOn
        watchersCount
        satisfactionRate
        ratingsCount
        officialSiteUrl
        officialSiteUrlEn
        wikipediaUrl
        wikipediaUrlEn
        twitterUsername
        twitterHashtag
        syobocalTid
        malAnimeId
        isNoEpisodes
        synopsis
        synopsisEn
        synopsisSource
        synopsisSourceEn
        image {
          internalUrl(size: "280x")
        }
        copyright
        trailers(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            title
            url
            internalImageUrl(size: "300x169")
          }
        }
        casts(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            name
            nameEn
            character {
              annictId
              name
              nameEn
            }
            person {
              annictId
              name
              nameEn
            }
          }
        }
        staffs(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            name
            nameEn
            role
            roleEn
            resource {
              __typename
              ... on Person {
                annictId
                name
                nameEn
              }
              ... on Organization {
                annictId
                name
                nameEn
              }
            }
          }
        }
        episodes(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            annictId
            numberText
            title
          }
        }
        programs {
          nodes {
            vodTitleCode
            vodTitleName
            channel {
              annictId
              name
            }
          }
        }
      }
    }
  }
`

export default async ({ workId }) => {
  const result = await client.query({ query: query, variables: { annictId: workId } })
  const node = result.data.works.nodes[0]
  console.log('node: ', node)
  const work = new Work(node)
  work.setSeason(node)
  work.setImage(node.image)
  work.setCasts(node.casts.nodes)
  work.setEpisodes(node.episodes.nodes)
  work.setTrailers(node.trailers.nodes)
  return work
}
