# frozen_string_literal: true

class IcsController < ApplicationV6Controller
  def show
    @user = User.only_kept.find_by!(username: params[:username])

    I18n.with_locale(@user.locale) do
      @slots = UserSlotsQuery.new(
        @user,
        Slot.only_kept.with_works(@user.animes_on(:wanna_watch, :watching).only_kept),
        watched: false
      ).call
        .where("started_at >= ?", Date.today.beginning_of_day)
        .where("started_at <= ?", 7.days.since.end_of_day)
        .where.not(episode_id: nil)

      @works = @user
        .animes_on(:wanna_watch, :watching)
        .only_kept
        .where.not(started_on: nil)

      render formats: :html, layout: false
    end
  end
end
