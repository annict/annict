# frozen_string_literal: true

module V3
  class WorkDetailQuery < V3::ApplicationQuery
    def initialize(work_id:)
      @work_id = work_id
    end

    def call
      build_object(execute(query_string, variables: { annict_id: work_id }))
    end

    private

    attr_reader :work_id

    def build_object(result)
      data = result.
        dig(:data, :works, :nodes).
        first
      return unless data

      attrs = data.slice(*WorkStruct.attribute_names)
      attrs[:image] = WorkImageStruct.new(data[:image])
      attrs[:trailers] = data.dig(:trailers, :nodes).map { |hash| TrailerStruct.new(hash.slice(*TrailerStruct.attribute_names)) }
      attrs[:casts] = data.dig(:casts, :nodes).map do |hash|
        CastStruct.new(
          character: CharacterStruct.new(hash[:character]),
          person: PersonStruct.new(hash[:person]),
        )
      end
      attrs[:staffs] = data.dig(:staffs, :nodes).map do |hash|
        StaffStruct.new(
          person: PersonStruct.new(hash[:resource]),
          organization: nil,
          role_text: hash[:role_text]
        )
      end
      attrs[:episodes] = data.dig(:episodes, :nodes).map { |hash| EpisodeStruct.new(hash.slice(*EpisodeStruct.attribute_names)) }

      WorkStruct.new(attrs)
    end

    def query_string
      <<~GRAPHQL
        query($annictId: Int!) {
          works(annictIds: [$annictId]) {
            nodes {
              id
              annictId
              title
              titleKana
              titleEn
              media
              seasonYear
              seasonName
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
              staffs(orderBy: { field: SORT_NUMBER, direction: ASC }) {
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
