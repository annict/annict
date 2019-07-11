# frozen_string_literal: true

module V3
  class FetchStatusQuery < V3::ApplicationQuery
    def initialize(user:, gql_work_id:)
      @user = user
      @gql_work_id = gql_work_id
    end

    def call
      build_object(execute(query_string, viewer: user))
    end

    private

    attr_reader :user, :gql_work_id

    def build_object(result)
      data = result.dig(:data, :node)
      WorkStruct.new(data.slice(*WorkStruct.attribute_names))
    end

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
