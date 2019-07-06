# frozen_string_literal: true

module V3
  class FetchStatusService
    def initialize(user:, gql_work_id:)
      @user = user
      @gql_work_id = gql_work_id
    end

    def call
      Canary::AnnictSchema.execute(query_string, context: { viewer: user })
    end

    private

    attr_reader :user, :gql_work_id

    def query_string
      <<~GRAPHQL
      {
        node(id: "#{gql_work_id}") {
          ... on Work {
            viewerStatusState
          }
        }
      }
      GRAPHQL
    end
  end
end
