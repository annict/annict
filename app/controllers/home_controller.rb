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
    @forum_posts = ForumPost.
      joins(:forum_category).
      merge(ForumCategory.with_slug(:site_news)).
      order(created_at: :desc).
      limit(5)

    tips = render_jb("home/_tips", tips: current_user.tips.unfinished.limit(3))

    latest_statuses = TrackableService.new(current_user).latest_statuses
    latest_status_data = render_jb "api/internal/latest_statuses/index",
      user: current_user,
      latest_statuses: latest_statuses

    activities = current_user.
      following_activities.
      order(id: :desc).
      includes(:recipient, trackable: :user, user: :profile).
      page(1)
    activity_data = render_jb("api/internal/activities/index",
      user: current_user,
      activities: activities)

    gon.push(tips: tips, latestStatusData: latest_status_data, activityData: activity_data)

    render :index
  end

  def load_i18n
    keys = {
      "messages._common.are_you_sure": nil,
      "messages.tracks.see_records": nil,
      "messages.tracks.skip_episode_confirmation": nil,
      "messages.tracks.tracked": nil
    }

    load_i18n_into_gon keys
  end
end
