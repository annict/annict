import gql from 'graphql-tag'

import client from '../client'
import { Work } from '../models'
import { ApplicationQuery } from './ApplicationQuery'

const query = gql`
  query($annictId: Int!) {
    works(annictIds: [$annictId]) {
      nodes {
        id
        annictId
        title
        titleKana
        titleEn
        localTitle
        media
        seasonName
        localSeasonName
        seasonSlug
        seasonYear
        localStartedOnLabel
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
        localSynopsis
        synopsisSource
        synopsisSourceEn
        localSynopsisSource
        viewerFinishedToWatch
        copyright
        image {
          internalUrl(size: "280x")
        }
        trailers(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            title
            url
            internalImageUrl(size: "300x169")
          }
        }
        casts(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            localAccuratedName
            character {
              annictId
              name
              nameEn
            }
            person {
              annictId
            }
          }
        }
        staffs(orderBy: { field: SORT_NUMBER, direction: ASC }) {
          nodes {
            localAccuratedName
            localRole
            resource {
              __typename
              ... on Person {
                annictId
              }
              ... on Organization {
                annictId
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
            vodTitleUrl
            channel {
              annictId
              name
            }
          }
        }
        workRecords(hasBody: true, filterByLocale: true, orderBy: { field: CREATED_AT, direction: DESC }, first: 10) {
          nodes {
            user {
              annictId
              username
              name
              avatarUrl(size: size50)
              isSupporter
            }
            record {
              annictId
              pageViewsCount
            }
            ratingAnimationState
            ratingMusicState
            ratingStoryState
            ratingCharacterState
            ratingOverallState
            body
            likesCount
            createdAt
            modifiedAt
            viewerDidLike
          }
        }
      }
    }
  }
`

export class FetchWorkQuery extends ApplicationQuery {
  private readonly workId: number

  public constructor({ workId }) {
    super()
    this.workId = workId
  }

  public async execute(): Promise<Work> {
    const result = await client.query({ query, variables: { annictId: this.workId } })
    const node = result.data.works.nodes[0]
    console.log('node: ', node)
    const work = new Work(node)
    work.setSeason(node)
    work.setImage(node.image)
    work.setTrailers(node.trailers.nodes)
    work.setCasts(node.casts.nodes)
    work.setStaffs(node.staffs.nodes)
    work.setEpisodes(node.episodes.nodes)
    work.setPrograms(node.programs.nodes)
    work.setWorkRecords(node.workRecords.nodes)
    return work
  }
}
