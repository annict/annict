# frozen_string_literal: true

module Db
  class TrailerPublishingsController < Db::ApplicationController
    include ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Trailer.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Trailer.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_trailer_list_path(create_resource.anime)
    end

    def after_destroyed_path
      db_trailer_list_path(destroy_resource.anime)
    end
  end
end
