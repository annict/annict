# typed: false
# frozen_string_literal: true

module Db
  class OrganizationPublishingsController < Db::ApplicationController
    include ResourcePublishable

    before_action :authenticate_user!

    private

    def create_resource
      @create_resource ||= Organization.without_deleted.unpublished.find(params[:id])
    end

    def destroy_resource
      @destroy_resource ||= Organization.without_deleted.published.find(params[:id])
    end

    def after_created_path
      db_organization_list_path
    end

    def after_destroyed_path
      db_organization_list_path
    end
  end
end
