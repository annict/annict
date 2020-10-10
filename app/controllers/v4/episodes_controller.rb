# frozen_string_literal: true

module V4
  class EpisodesController < V4::ApplicationController
    include AnimeSidebarDisplayable
    include EpisodeDisplayable

    def show
      set_page_category PageCategory::EPISODE

      load_episode_and_records(work_id: params[:work_id], episode_id: params[:id])
    end
  end
end
