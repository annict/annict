# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(update)

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

      def update
        library_entry = current_user.library_entries.find(params[:library_entry_id])
        program = if params[:program_id] != "no_select"
          library_entry.work.programs.only_kept.find(params[:program_id])
        end

        library_entry.update!(program_id: program&.id)

        head 204
      end
    end
  end
end
