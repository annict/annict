# frozen_string_literal: true

class WelcomeController < ApplicationController
  def show
    set_page_category PageCategory::GUEST_HOME

    @season_top_work = GuestTopPageService.season_top_work
    @season_works = GuestTopPageService.season_works
    @top_work = GuestTopPageService.top_work
    @works = GuestTopPageService.works
    @cover_image_work = GuestTopPageService.cover_image_work
  end
end
