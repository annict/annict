# frozen_string_literal: true

module Frame
  class AnimeController < Frame::ApplicationController
    def show
      @anime = Work.only_kept.find(params[:anime_id])
      @episodes = @anime.episodes.only_kept.order(:sort_number).page(params[:page]).per(10)
    end
  end
end
