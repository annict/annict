# frozen_string_literal: true

module V4
  module AnimeSidebarDisplayable
    def load_vod_channel_entities(anime:, anime_entity:)
      @vod_channel_entities = Rails.cache.fetch(anime_detail_vod_channels_cache_key(anime), expires_in: 3.hours) do
        AnimePage::VodChannelsRepository.new(graphql_client: graphql_client).execute(anime_entity: anime_entity)
      end

      @existing_vod_channel_entities = @vod_channel_entities.select { |vod_channel| vod_channel.programs.first.present? }
    end

    private

    def anime_detail_vod_channels_cache_key(anime)
      [
        "anime-detail",
        "vod-channels",
        anime.id,
        anime.updated_at.rfc3339
      ].freeze
    end
  end
end
