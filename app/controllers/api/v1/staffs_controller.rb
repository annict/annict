# frozen_string_literal: true

module API
  module V1
    class StaffsController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @staffs = Staff.without_deleted
        @staffs = API::V1::StaffIndexService.new(@staffs, @params).result
      end
    end
  end
end
