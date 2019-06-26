# frozen_string_literal: true

class FetchWorkDetailService
  def initialize(work_id:)
    @work_id = work_id
  end

  def call
    AnnictSchema.execute(query_string)
  end

  private

  attr_reader :work_id

  def query_string
    <<~GRAPHQL
      {
        searchWorks(annictIds: [#{work_id}]) {
          nodes {
            annictId
            title
            watchersCount
            satisfactionRate
            ratingsCount
            titleKana
            officialSiteUrl
            twitterUsername
            wikipediaUrl
            image {
              facebookOgImageUrl
            }
            casts {
              nodes {
                character {
                  annictId
                  name
                }
                person {
                  annictId
                  name
                }
              }
            }
            staffs {
              nodes {
                resource {
                  ... on Person {
                    annictId
                    name
                  }
                  ... on Organization {
                    annictId
                    name
                  }
                }
                roleText
              }
            }
            episodes(orderBy: { field: SORT_NUMBER, direction: ASC }) {
              nodes {
                annictId
                numberText
                title
              }
            }
          }
        }
      }
    GRAPHQL
  end
end
