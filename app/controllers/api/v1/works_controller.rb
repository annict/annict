# frozen_string_literal: true

module Api
  module V1
    class WorksController < Api::V1::ApplicationController
      before_action :prepare_params!, only: [:index]

      def index
        @works = Work.published.page(@params.page).per(@params.per_page)
        @works = @works.where(id: @params.filter_ids) if @params.filter_ids.present?
        @works = @works.by_season(@params.filter_season) if @params.filter_season.present?
        if @params.filter_title.present?
          @works = @works.search(title_or_title_kana_cont: @params.filter_title).result
        end
      end
    end
  end
end
