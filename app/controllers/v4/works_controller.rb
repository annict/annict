# frozen_string_literal: true

module V4
  class WorksController < V4::ApplicationController
    include V4::AnimeSidebarDisplayable

    def show
      set_page_category PageCategory::WORK

      anime = Work.only_kept.find(params[:anime_id])

      result = Rails.cache.fetch(work_detail_work_cache_key(anime), expires_in: 3.hours) {
        AnimePage::AnimeRepository.new(graphql_client: graphql_client).execute(anime_id: anime.id)
      }
      @anime_entity = result.anime_entity

      load_vod_channel_entities(anime: anime, anime_entity: @anime_entity)
    end

    private

    def work_detail_work_cache_key(anime)
      [
        "anime-detail",
        "anime",
        anime.id,
        anime.updated_at.rfc3339
      ].freeze
    end
  end
end
