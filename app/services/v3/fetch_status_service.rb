# frozen_string_literal: true

module V3
  class FetchStatusService
    def initialize(user:, work_id:)
      @user = user
      @work_id = work_id
    end

    def call
      AnnictSchema.execute(query_string, context: { viewer: user })
    end

    private

    attr_reader :user, :work_id

    def query_string
      <<~GRAPHQL
      {
        searchWorks(annictIds: [#{work_id}]) {
          nodes {
            viewerStatusState
          }
        }
      }
      GRAPHQL
    end
  end
end
