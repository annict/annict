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
    @tips = current_user.tips.unfinished.with_locale(current_user.locale)
    tips_data = render_jb("home/_tips", tips: @tips.limit(3))

    library_entries = TrackableService.new(current_user).library_entries
    library_entry_data = render_jb "api/internal/library_entries/index",
      user: current_user,
      library_entries: library_entries

    if device_pc?
      @forum_posts = ForumPost.
        joins(:forum_category).
        merge(ForumCategory.with_slug(:site_news))
      @forum_posts = localable_resources(@forum_posts).order(created_at: :desc).limit(5)

      activities = UserActivitiesQuery.new.call(
        activities: current_user.following_activities,
        user: current_user,
        page: 1
      )
      works = Work.where(id: activities.all.map(&:work_id))

      activity_data = render_jb("api/internal/activities/index",
        user: current_user,
        activities: activities,
        works: works)
    end

    gon.push(
      tipsData: tips_data,
      latestStatusData: library_entry_data,
      activityData: activity_data
    )

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
