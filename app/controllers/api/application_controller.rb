# frozen_string_literal: true

module Api
  class ApplicationController < ActionController::Base
    def ga_client
      @ga_client ||= Annict::Analytics::Client.new(request, current_user)
    end

    private

    def set_work
      @work = Work.find(params[:work_id])
    end

    def set_episode
      @episode = @work.episodes.find(params[:episode_id])
    end
  end
end
