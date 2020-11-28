# frozen_string_literal: true

module V4
  class TrackAnimeController < V4::ApplicationController
    before_action :authenticate_user!

    def show
      set_page_category PageCategory::TRACK_ANIME

      anime = Work.only_kept.find(params[:anime_id])

      @library_entry_entities = TrackPage::LibraryEntriesRepository.new(graphql_client: graphql_client(viewer: current_user)).execute
      @anime_entity, @page_info_entity = TrackAnimePage::AnimeRepository.new(graphql_client: graphql_client(viewer: current_user)).execute(
        anime_id: anime.id,
        pagination: Annict::Pagination.new(before: params[:before], after: params[:after], per: 500)
      )

      @form = AnimeRecordForm.new(anime_id: @anime_entity.id)
    end
  end
end
