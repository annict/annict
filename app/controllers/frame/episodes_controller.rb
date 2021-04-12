# frozen_string_literal: true

module Frame
  class EpisodesController < Frame::ApplicationController
    def show
      @episode = Episode.only_kept.find(params[:episode_id])
    end
  end
end
