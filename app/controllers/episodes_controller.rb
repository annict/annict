# frozen_string_literal: true

class EpisodesController < ApplicationController
  def index
    set_page_category Rails.configuration.page_categories.episode_list

    @work = Work.only_kept.find(params[:work_id])
    raise ActionController::RoutingError, "Not Found" if @work.no_episodes?

    @episodes = @work.episodes.only_kept.order(:sort_number)
  end
end
