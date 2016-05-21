# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include AnalyticsFilter

      before_action :set_paper_trail_whodunnit

      private

      def set_work
        @work = Work.find(params[:work_id])
      end

      def set_episode
        @episode = @work.episodes.find(params[:episode_id])
      end
    end
  end
end
