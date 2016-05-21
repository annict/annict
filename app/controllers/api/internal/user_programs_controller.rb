# frozen_string_literal: true

module Api
  module Internal
    class UserProgramsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index(page: nil)
        @programs = current_user.
          programs.
          unwatched.
          work_published.
          episode_published.
          where("started_at < ?", Date.tomorrow + 1.day + 5.hours).
          includes(:channel, :work, episode: [:work]).
          order(started_at: :desc).
          page(page)
      end
    end
  end
end
