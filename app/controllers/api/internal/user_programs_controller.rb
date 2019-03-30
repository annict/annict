# frozen_string_literal: true

module Api
  module Internal
    class UserProgramsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index
        @user = current_user
        @programs = current_user.programs.unwatched(params[:page], params[:sort])
      end
    end
  end
end
