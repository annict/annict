# frozen_string_literal: true

class HomeController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show
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
      latestStatusData: library_entry_data,
      activityData: activity_data
    )
  end

  private

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
