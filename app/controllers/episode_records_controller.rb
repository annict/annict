# frozen_string_literal: true

class EpisodeRecordsController < ApplicationController
  include V4::GraphqlRunnable

  before_action :authenticate_user!, only: %i(create edit update switch)

  def create
    @episode = Episode.only_kept.find(params[:episode_id])
    @work = @episode.work

    episode_record, err = CreateEpisodeRecordRepository.new(
      graphql_client: graphql_client(viewer: current_user)
    ).create(episode: @episode, params: episode_record_params)

    if err
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

      @is_spoiler = current_user.hide_episode_record_body?(@episode)
      @episode_record = @episode.episode_records.new(episode_record_params)
      @episode_record.errors.add(:mutation_error, err.message)
      @episode_record.setup_shared_sns(current_user)

      return render "/episodes/show"
    end

    flash[:notice] = t("messages.episode_records.created")

    redirect_to work_episode_path(@work, @episode)
  end

  def edit
    @episode_record = current_user.episode_records.only_kept.find_by(episode_id: params[:episode_id], record_id: params[:id])
    authorize @episode_record, :edit?
    @work = @episode_record.work
  end

  def update
    @episode_record = current_user.episode_records.only_kept.find_by(episode_id: params[:episode_id], record_id: params[:id])
    authorize @episode_record, :update?

    @episode_record.modify_body = true
    @episode_record.detect_locale!(:body)

    if @episode_record.update(episode_record_params)
      current_user.update_share_record_setting(@episode_record.share_to_twitter == "1")
      current_user.share_episode_record_to_twitter(@episode_record)

      path = record_path(current_user.username, @episode_record.record)
      redirect_to path, notice: t("messages._common.updated")
    else
      @work = @episode_record.work
      render :edit
    end
  end

  def switch
    episode = Episode.find(params[:episode_id])

    return redirect_to work_episode_path(episode.work, episode) unless params[:to].in?(Setting.display_option_record_list.values)

    current_user.setting.update_column(:display_option_record_list, params[:to])
    redirect_to work_episode_path(episode.work, episode)
  end

  def redirect
    url = case params[:provider]
    when "tw"
      EpisodeRecord.only_kept.find_by!(twitter_url_hash: params[:url_hash]).share_url_with_query(:twitter)
    when "fb"
      EpisodeRecord.only_kept.find_by!(facebook_url_hash: params[:url_hash]).share_url_with_query(:facebook)
    else
      root_path
    end

    redirect_to url, status: 301
  end

  private

  def episode_record_params
    params.require(:episode_record).permit(:episode_id, :body, :share_to_twitter, :rating_state)
  end
end
