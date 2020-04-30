# frozen_string_literal: true

module WorkDetail
  class VodChannelsRepository < ApplicationRepository
    def fetch(work:)
      result = graphql_client.execute(query)
      data = result.to_h.dig("data", "channels")

      data["nodes"].map do |node|
        VodChannelEntity.new(
          id: node["annictId"],
          name: node["name"],
          programs: work.programs.select do |program|
            program.channel.id == node["annictId"]
          end
        )
      end
    end

    private

    def query
      load_query "work_detail/fetch_vod_channels.graphql"
    end
  end
end
