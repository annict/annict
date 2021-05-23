# frozen_string_literal: true

module V4::Db
  class SeriesPublishingsController < V4::Db::ApplicationController
    include Db::ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Series.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Series.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_series_list_path
    end

    def after_destroyed_path
      db_series_list_path
    end
  end
end
