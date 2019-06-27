# frozen_string_literal: true

module V3
  class FetchWorkDetailService
    def initialize(work_id:)
      @work_id = work_id
    end

    def call
      AnnictSchema.execute(query_string, context: {
        admin: true
      })
    end

    private

    attr_reader :work_id

    def query_string
      <<~GRAPHQL
      {
        searchWorks(annictIds: [#{work_id}]) {
          nodes {
            id
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
              internalUrl(size: "280x")
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
end
