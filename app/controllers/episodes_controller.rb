# frozen_string_literal: true

class EpisodesController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def index
    @work = Work.published.find(params[:work_id])
    raise ActionController::RoutingError, "Not Found" if @work.no_episodes?

    @episodes = @work.episodes.published.order(:sort_number)

    return unless user_signed_in?

    store_page_params(work: @work)
  end

  def show
    @work = Work.published.find(params[:work_id])
    @episode = @work.episodes.published.find(params[:id])
    params[:locale_en] = locale_en?
    params[:locale_ja] = locale_ja?
    service = EpisodeRecordsListService.new(current_user, @episode, params)

    @all_episode_records = service.all_episode_records
    @all_comment_episode_records = service.all_comment_episode_records
    @friend_comment_episode_records = service.friend_comment_episode_records
    @my_episode_records = service.my_episode_records
    @selected_comment_episode_records = service.selected_comment_episode_records

    data = {
      recordsSortTypes: Setting.records_sort_type.options,
      currentRecordsSortType: current_user&.setting&.records_sort_type.presence || "created_at_desc"
    }
    gon.push(data)

    return unless user_signed_in?

    @is_spoiler = current_user.hide_episode_record_body?(@episode)
    @episode_record = @episode.episode_records.new
    @episode_record.setup_shared_sns(current_user)

    store_page_params(work: @work)
  end

  private

  def load_i18n
    keys = {
      "messages._common.are_you_sure": nil,
      "messages.components.mute_user_button.the_user_has_been_muted": nil
    }

    load_i18n_into_gon keys
  end
end
