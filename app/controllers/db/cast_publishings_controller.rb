# frozen_string_literal: true

module Db
  class CastPublishingsController < Db::ApplicationController
    include ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Cast.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Cast.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_cast_list_path(create_resource.anime)
    end

    def after_destroyed_path
      db_cast_list_path(destroy_resource.anime)
    end
  end
end
