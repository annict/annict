# frozen_string_literal: true

module V4
  class AnimeRecordsController < ApplicationController
    include AnimeSidebarDisplayable

    before_action :authenticate_user!, only: %i[create edit update destroy]

    def index
      set_page_category PageCategory::ANIME_RECORD_LIST
      @anime = Work.only_kept.find(params[:anime_id])
      load_on_anime_record_list
      @form = AnimeRecordForm.new(anime_id: @anime_entity.id)
    end

    def create
      @anime = Work.only_kept.find(params[:anime_id])
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

    def load_on_anime_record_list
      result = AnimeRecordListPage::AnimeRepository.new(graphql_client: graphql_client).execute(database_id: @anime.id)
      @anime_entity = result.anime_entity
      load_vod_channel_entities(anime: @anime, anime_entity: @anime_entity)
      @other_record_entities = @anime_entity.records

      if user_signed_in?
        result = AnimeRecordListPage::MyRecordsRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(anime_id: @anime_entity.id)
        @my_record_entities = result.record_entities

        result = AnimeRecordListPage::FollowingRecordsRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(anime_id: @anime_entity.id)
        @following_record_entities = result.record_entities

        @other_record_entities = current_user.filter_records(
          @other_record_entities, @my_record_entities + @following_record_entities
        )
      end
    end
  end
end
