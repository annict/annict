# frozen_string_literal: true

class TracksController < ApplicationController
  before_action :authenticate_user!, only: %i(show)
  before_action :load_i18n, only: %i(show)

  def show
    @latest_statuses = TrackableService.new(current_user).latest_statuses

    page_object = render_jb "api/internal/latest_statuses/index",
      user: current_user,
      latest_statuses: @latest_statuses
    gon.push(pageObject: page_object)
  end

  private

  def load_i18n
    keys = {
      "messages.tracks.skip_episode_confirmation": nil,
      "messages.tracks.see_records": nil,
      "messages.tracks.tracked": nil
    }

    load_i18n_into_gon keys
  end
end
