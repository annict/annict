# frozen_string_literal: true

class EpisodeItemsController < ApplicationController
  before_action :authenticate_user!, only: %i(new destroy)
  before_action :load_episode, only: %i(new destroy)
  before_action :set_page_object, only: %i(new)

  def new
    @work = @episode.work

    return unless browser.device.mobile?

    service = RecordsListService.new(current_user, @episode, params)
    @all_records = service.all_records
  end

  def destroy(id)
    item = @episode.items.published.find(id)
    episode_item = @episode.resource_items.find_by(item: item, user: current_user)

    episode_item.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_back fallback_location: work_episode_path(@episode.work, @episode)
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
