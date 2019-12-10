# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

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
