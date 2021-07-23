# frozen_string_literal: true

module Api
  module Internal
    class LibraryEntriesController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i[update]

      # POST /api/internal/library_entries
      # NOTE: `anime_ids` の量が大きいときがあるため、POSTで受け付けている
      def index
        return render(json: []) unless user_signed_in?
        return render(json: []) unless params[:anime_ids]

        status_kinds = current_user
          .library_entries
          .where(work_id: params[:anime_ids].split(","))
          .status_kinds

        render json: status_kinds
      end

      def update
        library_entry = current_user.library_entries.find(params[:library_entry_id])
        program = if params[:program_id] != "no_select"
          library_entry.anime.programs.only_kept.find(params[:program_id])
        end

        library_entry.update!(program_id: program&.id)

        head 204
      end
    end
  end
end
