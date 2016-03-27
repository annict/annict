# frozen_string_literal: true

class HomeController < ApplicationController
  def index
    return render :index, layout: layout if user_signed_in?

    @season_top_work = GuestTopPageService.season_top_work
    @season_works = GuestTopPageService.season_works
    @top_work = GuestTopPageService.top_work
    @works = GuestTopPageService.works
    @cover_image_work = GuestTopPageService.cover_image_work
    @activities = GuestTopPageService.activities unless browser.mobile?

    render :index_guest, layout: "v2/application"
  end

  private

  def layout
    return "v1/application" if browser.mobile?
    "v2/application"
  end
end
