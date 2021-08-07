# frozen_string_literal: true

module Fragment
  class AnimeRecordsController < Fragment::ApplicationController
    include AnimeRecordListSettable

    before_action :authenticate_user!, only: %i[index]

    def index
      anime = Anime.only_kept.find(params[:anime_id])

      set_anime_record_list(anime)
    end
  end
end
