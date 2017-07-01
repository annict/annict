# frozen_string_literal: true

class EpisodeItemsController < ApplicationController
  before_action :authenticate_user!, only: %i(new)
  before_action :load_episode, only: %i(new)
  before_action :set_page_object, only: %i(new)

  def new
    @work = @episode.work

    return unless browser.device.mobile?

    service = RecordsListService.new(current_user, @episode, params)
    @all_records = service.all_records
  end

  private

  def load_episode
    @episode = Episode.published.find(params[:episode_id])
  end

  def set_page_object
    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @episode.work
  end
end
