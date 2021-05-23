# frozen_string_literal: true

module V4::Db
  class TrailerPublishingsController < V4::Db::ApplicationController
    include Db::ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Trailer.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Trailer.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_trailer_list_path(create_resource.work)
    end

    def after_destroyed_path
      db_trailer_list_path(destroy_resource.work)
    end
  end
end
