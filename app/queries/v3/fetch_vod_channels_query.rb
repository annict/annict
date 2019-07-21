# frozen_string_literal: true

module V3
  class FetchVodChannelsQuery < V3::ApplicationQuery
    def call
      build_object(execute(query_string))
    end

    private

    def build_object(result)
      nodes = result.dig(:data, :channels, :nodes)
      nodes.map { |node| ChannelStruct.new(node.slice(*ChannelStruct.attribute_names)) }
    end

      def query_string
      <<~GRAPHQL
        query {
          channels(isVod: true) {
            nodes {
              annictId
              name
            }
          }
        }
      GRAPHQL
    end
  end
end
