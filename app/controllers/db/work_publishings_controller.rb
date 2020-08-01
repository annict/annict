# frozen_string_literal: true

module Db
  class WorkPublishingsController < Db::ApplicationController
    include V4::ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Anime.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Anime.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_work_list_path
    end

    def after_destroyed_path
      db_work_list_path
    end
  end
end
