# frozen_string_literal: true

module My
  class AnimeSidebarController < My::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(show)

    def show
      @user = current_user
      @anime = Anime.only_kept.find(params[:anime_id])
    end
  end
end
