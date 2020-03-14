# frozen_string_literal: true

module Api
  module Internal
    class SlotsSortTypesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def update
        current_user.setting.update_column(
          :slots_sort_type, permitted[:slots_sort_type]
        )
        # To clear cache
        current_user.channel_works.update_all(updated_at: Time.now)
        head 204
      end

      private

      def permitted
        params.permit(:slots_sort_type)
      end
    end
  end
end
