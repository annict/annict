# typed: false
# frozen_string_literal: true

module Db
  class StaffPublishingsController < Db::ApplicationController
    include ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Staff.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Staff.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_staff_list_path(create_resource.work)
    end

    def after_destroyed_path
      db_staff_list_path(destroy_resource.work)
    end
  end
end
