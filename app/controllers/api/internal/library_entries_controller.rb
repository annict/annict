# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(show skip_episode)

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

      def show
        @library_entry = current_user.library_entries.find_by(work_id: params[:work_id])
        @user = current_user
      end

      def skip_episode
        @library_entry = LibraryEntry.find(params[:library_entry_id])
        @library_entry.append_episode!(@library_entry.next_episode)
        @user = current_user
        render :show
      end
    end
  end
end
