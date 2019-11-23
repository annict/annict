# frozen_string_literal: true

class IcsController < ApplicationController
  def show
    @user = User.without_deleted.find_by!(username: params[:username])

    I18n.with_locale(@user.locale) do
      @slots = @user.
        slots.
        unwatched_all.
        work_published.
        episode_published.
        includes(:work, :episode, :channel).
        where("started_at >= ?", Date.today.beginning_of_day).
        where("started_at <= ?", 7.days.since.end_of_day)
      @works = @user.
        works.
        wanna_watch_and_watching.
        where.not(started_on: nil)

      render layout: false
    end
  end
end
