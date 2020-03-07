# frozen_string_literal: true

module DB
  class WorksController < DB::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update destroy)

    def index
      @work_conn = WorkRepository.new(viewer: current_user).resent
    end

    def season
      slug = params[:slug].presence || ENV.fetch("ANNICT_CURRENT_SEASON")
      @works = Work.by_season(slug).order(id: :desc).page(params[:page])
      render :index
    end

    def resourceless
      name = params[:name].presence || "episode"
      @works = case name
      when "episode" then Work.where(auto_episodes_count: 0, no_episodes: false)
      when "item" then Work.image_not_attached
      end
      @works = @works.order(watchers_count: :desc).page(params[:page])
      render :index
    end

    def new
      @work = Work.new
      authorize @work, :new?
    end

    def create
      @work = Work.new(work_params)
      @work.user = current_user
      authorize @work, :create?

      return render(:new) unless @work.valid?
      @work.save_and_create_activity!

      redirect_to edit_db_work_path(@work), notice: t("resources.work.created")
    end

    def edit
      @work = Work.find(params[:id])
      authorize @work, :edit?
    end

    def update
      @work = Work.find(params[:id])
      authorize @work, :update?

      @work.attributes = work_params
      @work.user = current_user

      return render(:edit) unless @work.valid?
      @work.save_and_create_activity!

      redirect_to edit_db_work_path(@work), notice: t("resources.work.updated")
    end

    def hide
      @work = Work.find(params[:id])
      authorize @work, :hide?

      @work.soft_delete_with_children

      flash[:notice] = t("resources.work.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @work = Work.find(params[:id])
      authorize @work, :destroy?

      @work.destroy

      flash[:notice] = t("resources.work.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @work = Work.find(params[:id])
      @activities = @work.db_activities.order(id: :desc)
      @comment = @work.db_comments.new
    end

    private

    def work_params
      params.require(:work).permit(
        :title, :title_kana, :title_alter, :title_en, :title_alter_en, :media, :official_site_url,
        :official_site_url_en, :wikipedia_url, :wikipedia_url_en, :twitter_username,
        :twitter_hashtag, :sc_tid, :mal_anime_id, :number_format_id, :synopsis,
        :synopsis_source, :synopsis_en, :synopsis_source_en, :season_year, :season_name,
        :manual_episodes_count, :start_episode_raw_number, :no_episodes, :started_on, :ended_on
      )
    end
  end
end
