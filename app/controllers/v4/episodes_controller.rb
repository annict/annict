# frozen_string_literal: true

module V4
  class EpisodesController < V4::ApplicationController
    include AnimeSidebarDisplayable

    def show
      set_page_category PageCategory::EPISODE

      anime = Work.only_kept.find(params[:work_id])
      episode = anime.episodes.only_kept.find(params[:id])

      @episode_entity = Rails.cache.fetch(episode_cache_key(episode), expires_in: 3.days) do
        EpisodePage::EpisodeRepository.new(graphql_client: graphql_client).execute(database_id: episode.id)
      end

      load_vod_channel_entities(anime: anime, anime_entity: @episode_entity.anime)

      @form = EpisodeRecordForm.new(episode_id: @episode_entity.id)

      if user_signed_in?
        @my_record_entities = Rails.cache.fetch(my_records_cache_key(current_user), expires_in: 1.day) do
          EpisodePage::MyRecordsRepository.new(graphql_client: graphql_client(viewer: current_user)).execute(episode_id: @episode_entity.id)
        end
      end
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

    def my_records_cache_key(user)
      [
        "episode-page",
        "my-records",
        user.id,
        user.last_record_watched_at.rfc3339
      ].freeze
    end
  end
end
