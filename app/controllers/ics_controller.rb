# frozen_string_literal: true

class IcsController < ApplicationV6Controller
  def show
    @user = User.only_kept.find_by!(username: params[:username])
    library_entries = @user.library_entries.wanna_watch_and_watching.where.not(program_id: nil)

    I18n.with_locale(@user.locale) do
      @slots = Slot
        .only_kept
        .where(program_id: library_entries.pluck(:program_id))
        .where("started_at >= ?", Date.today.beginning_of_day)
        .where("started_at <= ?", 7.days.since.end_of_day)
        .where.not(episode_id: nil)
        .where.not(episode_id: library_entries.pluck(:watched_episode_ids).flatten)

      @works = @user
        .works_on(:wanna_watch, :watching)
        .only_kept
        .where.not(started_on: nil)

      render formats: :html, layout: false
    end
  end
end
