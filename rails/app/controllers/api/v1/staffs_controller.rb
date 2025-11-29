# typed: false
# frozen_string_literal: true

module Api
  module V1
    class StaffsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @staffs = Staff.only_kept
        @staffs = Deprecated::Api::V1::StaffIndexService.new(@staffs, @params).result
      end
    end
  end
end
