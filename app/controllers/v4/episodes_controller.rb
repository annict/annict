# frozen_string_literal: true

module V4
  class EpisodesController < V4::ApplicationController
    include AnimeSidebarDisplayable

    def show
      set_page_category PageCategory::EPISODE

      anime = Work.only_kept.find(params[:work_id])
      episode = anime.episodes.only_kept.find(params[:id])

      @episode_entity = Rails.cache.fetch(episode_cache_key(episode), expires_in: 3.hours) do
        EpisodePage::EpisodeRepository.new(graphql_client: graphql_client).execute(episode_id: episode.id)
      end

      load_vod_channel_entities(anime: anime, anime_entity: @episode_entity.anime)

      @form = EpisodeRecordForm.new(episode_id: @episode_entity.id)
    end

    private

    def episode_cache_key(episode)
      [
        "episode-page",
        "episode",
        episode.id,
        episode.updated_at.rfc3339,
        episode.last_record_watched_at.rfc3339
      ].freeze
    end
  end
end
