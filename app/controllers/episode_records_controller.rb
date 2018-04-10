# frozen_string_literal: true

class EpisodeRecordsController < ApplicationController
  permits :episode_id, :comment, :shared_twitter, :shared_facebook, :rating_state, model_name: "Record"

  before_action :authenticate_user!, only: %i(create switch)

  def create(episode_id, record)
    @episode = Episode.published.find(episode_id)
    @work = @episode.work
    @record = @episode.records.new(record)
    @record.work = @work
    ga_client.page_category = params[:page_category]

    service = NewRecordService.new(current_user, @record)
    service.page_category = params[:page_category]
    service.ga_client = ga_client
    service.via = "web"

    begin
      service.save!
      flash[:notice] = t("messages.records.created")
      redirect_to work_episode_path(@work, @episode)
    rescue
      params[:locale_en] = locale_en?
      params[:locale_ja] = locale_ja?
      service = RecordsListService.new(current_user, @episode, params)

      @all_records = service.all_records
      @all_comment_records = service.all_comment_records
      @friend_comment_records = service.friend_comment_records
      @my_records = service.my_records
      @selected_comment_records = service.selected_comment_records

      data = {
        recordsSortTypes: Setting.records_sort_type.options,
        currentRecordsSortType: current_user&.setting&.records_sort_type.presence || "created_at_desc"
      }
      gon.push(data)

      store_page_params(work: @work)

      @is_spoiler = current_user.hide_record?(@record)

      render "/episodes/show"
    end
  end

  def switch(episode_id, to)
    episode = Episode.published.find(episode_id)

    return redirect_to work_episode_path(episode.work, episode) unless to.in?(Setting.display_option_record_list.values)

    current_user.setting.update_column(:display_option_record_list, to)
    redirect_to work_episode_path(episode.work, episode)
  end
end
