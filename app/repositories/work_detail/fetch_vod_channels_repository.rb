# frozen_string_literal: true

module WorkDetail
  class FetchVodChannelsRepository < ApplicationRepository
    def fetch(work:)
      result = execute
      data = result.to_h.dig("data", "channels")

      data["nodes"].map do |node|
        VodChannelEntity.new(
          id: node["databaseId"],
          name: node["name"],
          programs: work.programs.select do |program|
            program.channel.id == node["databaseId"]
          end
        )
      end
    end
  end
end
