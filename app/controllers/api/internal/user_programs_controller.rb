# frozen_string_literal: true

module Api
  module Internal
    class UserProgramsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index(page: nil, sort: nil)
        @programs = current_user.
          programs.
          unwatched.
          work_published.
          episode_published.
          where("started_at < ?", Date.tomorrow + 1.day + 5.hours).
          includes(:channel, :work, episode: [:work]).
          order(started_at: sort_type(sort)).
          page(page)
      end

      private

      def sort_type(sort)
        return :asc if sort == "started_at_asc"
        return :desc if sort == "started_at_desc"
        :desc
      end
    end
  end
end
