# frozen_string_literal: true

class TracksController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show
    library_entries = TrackableService.new(current_user).library_entries
    library_entry_data = render_jb "api/internal/library_entries/index",
      user: current_user,
      library_entries: library_entries

    gon.push(
      latestStatusData: library_entry_data
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
