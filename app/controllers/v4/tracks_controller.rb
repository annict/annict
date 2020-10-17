# frozen_string_literal: true

module V4
  class TracksController < V4::ApplicationController
    before_action :authenticate_user!

    def show
      set_page_category PageCategory::TRACK

      @slot_entities = TrackPage::SlotsRepository.new(graphql_client: graphql_client(viewer: current_user)).execute
    end
  end
end
