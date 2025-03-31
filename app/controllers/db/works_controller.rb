# typed: false
# frozen_string_literal: true

module Db
  class WorksController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @is_no_episodes = search_params[:no_episodes] == "1"
      @is_no_image = search_params[:no_image] == "1"
      @is_no_release_season = search_params[:no_release_season] == "1"
      @is_no_slots = search_params[:no_slots] == "1"
      @season_slugs = search_params[:season_slugs]

      @works = Work.without_deleted.preload(:work_image)
      @works = @works.with_no_episodes if @is_no_episodes
      @works = @works.with_no_image if @is_no_image
      @works = @works.with_no_season if @is_no_release_season
      @works = @works.with_no_slots if @is_no_slots
      @works = @works.by_seasons(@season_slugs) if @season_slugs.present?
      @works = @works.order(id: :desc).page(params[:page]).per(100)
    end

    def new
      @work = Work.new
      authorize @work
    end

    def create
      @work = Work.new(work_params)
      @work.user = current_user
      authorize @work

      return render(:new, status: :unprocessable_entity) unless @work.valid?

      @work.save_and_create_activity!

      redirect_to db_edit_work_path(@work), notice: t("resources.work.created")
    end

    def edit
      @work = Work.without_deleted.find(params[:id])
      authorize @work
    end

    def update
      @work = Work.without_deleted.find(params[:id])
      authorize @work

      @work.attributes = work_params
      @work.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @work.valid?

      @work.save_and_create_activity!

      redirect_to db_edit_work_path(@work), notice: t("resources.work.updated")
    end

    def destroy
      @work = Work.without_deleted.find(params[:id])
      authorize @work

      @work.destroy_in_batches

      redirect_back(
        fallback_location: db_work_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def search_params
      params.permit(:commit, :no_episodes, :no_image, :no_release_season, :no_slots, season_slugs: [])
    end

    def work_params
      params.require(:work).permit(
        :title, :title_kana, :title_alter, :title_en, :title_alter_en, :media, :official_site_url,
        :official_site_url_en, :wikipedia_url, :wikipedia_url_en, :twitter_username,
        :twitter_hashtag, :sc_tid, :mal_anime_id, :number_format_id, :synopsis,
        :synopsis_source, :synopsis_en, :synopsis_source_en, :season_year, :season_name,
        :manual_episodes_count, :start_episode_raw_number, :started_on, :ended_on
      )
    end
  end
end
