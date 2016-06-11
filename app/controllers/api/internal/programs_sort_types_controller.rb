# frozen_string_literal: true

module Api
  module Internal
    class ProgramsSortTypesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def update
        current_user.setting.update_column(
          :programs_sort_type, permitted[:programs_sort_type]
        )
        render status: 204, nothing: true
      end

      private

      def permitted
        params.permit(:programs_sort_type)
      end
    end
  end
end
