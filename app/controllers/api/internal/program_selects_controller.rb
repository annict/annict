# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class ProgramSelectsController < Api::Internal::ApplicationController
      def create
        return head(:unauthorized) unless user_signed_in?

        work = Work.only_kept.find(params[:work_id])
        program = params[:program_id] == "0" ? nil : work.programs.only_kept.find(params[:program_id])
        current_user.save_program_to_library_entry!(work, program)

        head 204
      end
    end
  end
end
