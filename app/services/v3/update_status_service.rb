# frozen_string_literal: true

module V3
  class UpdateStatusService
    def initialize(user:, gql_work_id:, status_kind:)
      @user = user
      @gql_work_id = gql_work_id
      @status_kind = status_kind
    end

    def call
      Canary::AnnictSchema.execute(query_string, context: {
        viewer: user,
        writable: true
      })
    end

    private

    attr_reader :user, :gql_work_id, :status_kind

    def query_string
      <<~GRAPHQL
        mutation {
          statusUpdate(input: { workId: #{gql_work_id}, state: #{status_kind.upcase} }) {
            work {
              id
            }
          }
        }
      GRAPHQL
    end
  end
end
