# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      # POST /api/internal/library_entries
      # NOTE: `work_ids` の量が大きいときがあるため、POSTで受け付けている
      def index
        return render(json: []) unless user_signed_in?
        return render(json: []) unless params[:work_ids]

        status_kinds = current_user
          .library_entries
          .where(work_id: params[:work_ids].split(","))
          .status_kinds

        render json: status_kinds
      end
    end
  end
end
