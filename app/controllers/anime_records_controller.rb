# frozen_string_literal: true

class AnimeRecordsController < ApplicationV6Controller
  before_action :authenticate_user!, only: %i[create edit update destroy]

  def index
    set_page_category PageCategory::ANIME_RECORD_LIST

    @anime = Anime.only_kept.find(params[:anime_id])
    @programs = @anime.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))

    @form = Forms::AnimeRecordForm.new(anime: @anime)

    records = @anime
      .records
      .only_kept
      .eager_load(:anime_record, user: %i[gumroad_subscriber profile setting])
      .merge(AnimeRecord.only_kept.with_body.order_by_rating(:desc).order(created_at: :desc))
    @my_records = @following_records = []

    if user_signed_in?
      @my_records = current_user.records.only_kept.where(anime: @anime)
      @following_records = records.merge(current_user.followings)
      @other_records = records
        .where.not(user: [current_user, *current_user.followings])
        .page(params[:page])
        .per(100)
        .without_count
    else
      @other_records = records
        .page(params[:page])
        .per(100)
        .without_count
    end
  end

  def create
    @anime = Anime.only_kept.find(params[:anime_id])
    @form = AnimeRecordForm.new(anime_record_form_params)

    if @form.invalid?
      set_page_category PageCategory::ANIME_RECORD_LIST
      load_on_anime_record_list
      return render :index
    end

    _, err = CreateAnimeRecordRepository.new(graphql_client: graphql_client(viewer: current_user)).execute(form: @form)

    if err
      @form.errors.full_messages = [err.message]
      load_on_anime_record_list
      return render :index
    end

    flash[:notice] = t("messages._common.post")

    redirect_to anime_record_list_path(anime_id: @anime.id)
  end

  def edit
    set_page_category Rails.configuration.page_categories.record_edit

    @record = current_user.records.find(params[:id])
    @work_record = current_user
      .work_records
      .only_kept
      .where(work_id: params[:work_id])
      .find(@record.work_record&.id)
    @work = @work_record.work
    authorize @work_record, :edit?
  end

  def update
    @record = current_user.records.find(params[:id])
    @work_record = current_user
      .work_records
      .only_kept
      .where(work_id: params[:work_id])
      .find(@record.work_record&.id)
    @work = @work_record.work
    authorize @work_record, :update?

    @work_record.attributes = work_record_params
    @work_record.detect_locale!(:body)
    @work_record.modified_at = Time.now

    begin
      ActiveRecord::Base.transaction do
        @work_record.save!
        current_user.update_share_record_setting(@work_record.share_to_twitter == "1")
        current_user.share_work_record_to_twitter(@work_record)
      end
      flash[:notice] = t("messages._common.updated")
      redirect_to record_path(@work_record.user.username, @work_record.record)
    rescue
      render :edit
    end
  end

  private

  def anime_record_form_params
    params.require(:anime_record_form).permit(
      :anime_id, :comment, :rating_animation,
      :rating_music, :rating_story, :rating_character, :rating_overall,
      :share_to_twitter
    )
  end
end
