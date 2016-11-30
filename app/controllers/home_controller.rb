# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index
    if user_signed_in?
      gon.tips = render_jb("home/_tips", tips: current_user.tips.unfinished.limit(3))
      return render :index
    end

    @season_top_work = GuestTopPageService.season_top_work
    @season_works = GuestTopPageService.season_works
    @top_work = GuestTopPageService.top_work
    @works = GuestTopPageService.works
    @cover_image_work = GuestTopPageService.cover_image_work

    render :index_guest
  end

  private

  def load_i18n
    keys = {
      "messages.registrations.new.username_preview": {
        mobile: "messages.registrations.new.username_preview_mobile"
      },
      "messages.are_you_sure": nil
    }

    load_i18n_into_gon keys
  end
end
