# frozen_string_literal: true

module V4
  class WorksController < V4::ApplicationController
    def show
      set_page_category Rails.configuration.page_categories.work_detail

      work = Work.only_kept.find(params[:id])

      @work_entity = Rails.cache.fetch(work_detail_work_cache_key(work), expires_in: 3.hours) do
        AnimeDetail::AnimeRepository.new(graphql_client: graphql_client).execute(anime_id: work.id)
      end

      @vod_channels = Rails.cache.fetch(work_detail_vod_channels_cache_key(work), expires_in: 3.hours) do
        AnimeDetail::VodChannelsRepository.new(graphql_client: graphql_client).execute(anime_entity: @work_entity)
      end
      @existing_vod_channels = @vod_channels.select { |vod_channel| vod_channel.programs.first.present? }
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

    def work_detail_vod_channels_cache_key(work)
      [
        "work-detail",
        "vod-channels",
        work.id,
        work.updated_at.rfc3339
      ].freeze
    end
  end
end
