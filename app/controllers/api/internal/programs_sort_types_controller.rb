# frozen_string_literal: true

module Api
  module Internal
    class ProgramsSortTypesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def update
        current_user.setting.update_column(
          :programs_sort_type, permitted[:programs_sort_type]
        )
        head 204
      end

      private

      def permitted
        params.permit(:programs_sort_type)
      end
    end
  end
end
