# frozen_string_literal: true

module Api
  module Internal
    class ProgramSelectsController < Api::Internal::ApplicationController
      def create
        return head(:unauthorized) unless user_signed_in?

        anime = Anime.only_kept.find(params[:anime_id])
        program = anime.programs.only_kept.find(params[:program_id])

        program.save_library_entry!(current_user)

        head 204
      end
    end
  end
end
