# frozen_string_literal: true

class TracksController < ApplicationController
  before_action :authenticate_user!, only: %i(show)
  before_action :load_i18n, only: %i(show)

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
