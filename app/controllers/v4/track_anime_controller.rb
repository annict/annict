# frozen_string_literal: true

module V4
  class TrackAnimeController < V4::ApplicationController
    before_action :authenticate_user!

    def show
      set_page_category PageCategory::TRACK_ANIME

      anime = Work.only_kept.find(params[:id])

      @library_entry_entities = TrackPage::LibraryEntriesRepository.new(graphql_client: graphql_client(viewer: current_user)).execute
      @anime_entity = TrackAnimePage::AnimeRepository.new(graphql_client: graphql_client).execute(anime_id: anime.id)
    end
  end
end
