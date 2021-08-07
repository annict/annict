# frozen_string_literal: true

module Api
  module V1
    class OrganizationsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @organizations = Organization.only_kept
        @organizations = Api::V1::OrganizationIndexService.new(@organizations, @params).result
      end
    end
  end
end
