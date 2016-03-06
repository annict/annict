# frozen_string_literal: true

class HomeController < ApplicationController
  def index
    if user_signed_in?
      render :index, layout: "v1/application"
    else
      @season_top_work = GuestTopPageService.season_top_work
      @season_works = GuestTopPageService.season_works
      @top_work = GuestTopPageService.top_work
      @works = GuestTopPageService.works
      @cover_image_work = GuestTopPageService.cover_image_work
      @activities = GuestTopPageService.activities

      render :index_guest, layout: "v2/application"
    end
  end
end
