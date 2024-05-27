# typed: false
# frozen_string_literal: true

module Db
  class SeriesWorkPublishingsController < Db::ApplicationController
    include ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= SeriesWork.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= SeriesWork.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_series_work_list_path(create_resource.series_id)
    end

    def after_destroyed_path
      db_series_work_list_path(destroy_resource.series_id)
    end
  end
end
