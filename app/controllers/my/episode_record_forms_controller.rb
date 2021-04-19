# frozen_string_literal: true

module My
  class EpisodeRecordFormsController < My::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(show)

    def show
      episode = Episode.only_kept.find(params[:episode_id])
      @form = EpisodeRecordForm.new(episode: episode)
    end
  end
end
