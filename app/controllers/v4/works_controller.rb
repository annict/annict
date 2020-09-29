# frozen_string_literal: true

module V4
  class WorksController < V4::ApplicationController
    include AnimeSidebarDisplayable

    def show
      set_page_category Rails.configuration.page_categories.work_detail

      work = Work.only_kept.find(params[:id])

      @work_entity = Rails.cache.fetch(work_detail_work_cache_key(work), expires_in: 3.hours) do
        AnimePage::AnimeRepository.new(graphql_client: graphql_client).execute(anime_id: work.id)
      end

      load_vod_channel_entities(work: work, work_entity: @work_entity)
    end

    private

    def work_detail_work_cache_key(work)
      [
        "work-detail",
        "work",
        work.id,
        work.updated_at.rfc3339
      ].freeze
    end
  end
end
