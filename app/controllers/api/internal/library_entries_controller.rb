# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        library_entries = current_user.
          library_entries.
          joins(:status).
          select("library_entries.work_id, statuses.kind as status_kind").
          map do |library_entry|
            {
              work_id: library_entry.work_id,
              status_kind:  Status.kind.find_value(library_entry.status_kind)
            }
          end

        render json: library_entries
      end
    end
  end
end
