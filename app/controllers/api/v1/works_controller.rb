# frozen_string_literal: true

module Api
  module V1
    class WorksController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @works = Work.published
        @works = @works.where(id: @params.filter_ids) if @params.filter_ids.present?
        @works = @works.by_season(@params.filter_season) if @params.filter_season.present?
        if @params.filter_title.present?
          @works = @works.search(title_or_title_kana_cont: @params.filter_title).result
        end
        @works = @works.order(id: @params.sort_id) if @params.sort_id.present?
        if @params.sort_season.present?
          @works = @works.order_by_season(@params.sort_season)
        end
        if @params.sort_watchers_count.present?
          @works = @works.order(watchers_count: @params.sort_watchers_count)
        end
        @works = @works.page(@params.page).per(@params.per_page)
      end
    end
  end
end
