# frozen_string_literal: true

class RecordsController < ApplicationV6Controller
  include Pundit

  before_action :authenticate_user!, only: %i[update]

  def index
    set_page_category PageCategory::RECORD_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @dates = @user.records.only_kept.group_by_month(:created_at).count.to_a.reverse.to_h

    @records = @user
      .records
      .preload(:anime_record, anime: :anime_image, episode_record: :episode)
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count
    @records = @records.by_month(params[:month], year: params[:year]) if params[:month] && params[:year]
    @anime_ids = @records.pluck(:work_id)
  end

  def update
    user = User.only_kept.find_by!(username: params[:username])
    @record = current_user.records.only_kept.find_by!(id: params[:record_id], user_id: user.id)

    authorize @record, :update?

    if @record.episode_record?
      @form = EpisodeRecordForm.new(episode_record_form_params)
      @form.record = @record
      @form.episode = @record.episode_record.episode

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      EpisodeRecordUpdater.new(user: current_user, form: @form).call
    end

    head 204
  end

  private

  def episode_record_form_params
    params.required(:episode_record_form).permit(:comment, :rating, :share_to_twitter)
  end
end
