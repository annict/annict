# frozen_string_literal: true

module API
  module V1
    class OrganizationsController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @organizations = Organization.without_deleted
        @organizations = API::V1::OrganizationIndexService.new(@organizations, @params).result
      end
    end
  end
end
