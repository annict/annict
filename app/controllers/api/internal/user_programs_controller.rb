# frozen_string_literal: true

module Api
  module Internal
    class UserProgramsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index(page: nil, sort: nil)
        @user = current_user
        @programs = current_user.programs.unwatched(page, sort)
      end
    end
  end
end
