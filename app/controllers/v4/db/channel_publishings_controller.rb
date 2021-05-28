# frozen_string_literal: true

module V4::Db
  class ChannelPublishingsController < V4::Db::ApplicationController
    include V4::Db::ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Channel.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Channel.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_channel_list_path
    end

    def after_destroyed_path
      db_channel_list_path
    end
  end
end
