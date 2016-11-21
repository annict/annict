# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include ControllerCommon

      private

      def load_work
        @work = Work.published.find(params[:work_id])
      end

      def load_episode
        @episode = @work.episodes.find(params[:episode_id])
      end
    end
  end
end
