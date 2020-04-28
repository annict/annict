# frozen_string_literal: true

module Api
  module Internal
    class ViewersController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(show)

      def show
        library_entries = current_user.
          library_entries.
          joins(:status).
          select("library_entries.work_id, statuses.kind as status_kind")

        library_entries = library_entries.map do |library_entry|
          {
            work_id: library_entry.work_id,
            status_kind:  Status.kind.find_value(library_entry.status_kind)
          }
        end

        render json: {
          library_entries: library_entries
        }
      end
    end
  end
end
