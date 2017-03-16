# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index
    return index_member if user_signed_in?

    @season_top_work = GuestTopPageService.season_top_work
    @season_works = GuestTopPageService.season_works
    @top_work = GuestTopPageService.top_work
    @works = GuestTopPageService.works
    @cover_image_work = GuestTopPageService.cover_image_work

    render :index_guest
  end

  private

  def index_member
    tips = render_jb("home/_tips", tips: current_user.tips.unfinished.limit(3))
    activities = current_user.
      following_activities.
      order(id: :desc).
      includes(:recipient, trackable: :user, user: :profile).
      page(1)
    page_object = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities)

    gon.push(tips: tips, pageObject: page_object)

    render :index
  end

  def load_i18n
    keys = {
      "messages._common.are_you_sure": nil
    }

    load_i18n_into_gon keys
  end
end
