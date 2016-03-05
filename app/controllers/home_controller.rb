# frozen_string_literal: true

class HomeController < ApplicationController
  def index
    if user_signed_in?
      render :index, layout: "v1/application"
    else
      @season_top_work = Work.
        published.
        by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).
        order(watchers_count: :desc).
        first

      @season_works = Work.
        published.
        by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).
        where.not(id: @season_top_work.id).
        order(watchers_count: :desc).
        limit(8)

      cover_image_work_id = [@season_top_work.id, @season_works.pluck(:id)].flatten.sample
      @cover_image_work = Work.find(cover_image_work_id)

      @activities = Activity.where(action: ["checkins.create", "statuses.create"]).
        order(id: :desc).
        includes(:recipient, trackable: :user, user: :profile).
        limit(30)

      render :index_guest, layout: "v2/application"
    end
  end
end
