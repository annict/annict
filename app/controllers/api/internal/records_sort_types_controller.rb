# frozen_string_literal: true

module Api
  module Internal
    class RecordsSortTypesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def update
        current_user.setting.update_column(
          :records_sort_type, permitted[:records_sort_type]
        )
        head 204
      end

      private

      def permitted
        params.permit(:records_sort_type)
      end
    end
  end
end
