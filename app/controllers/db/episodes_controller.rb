# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @episodes = @work.episodes.
        includes(:prev_episode).
        order(sort_number: :desc).
        page(params[:page])
    end

    def new
      @work = Work.find(params[:work_id])
      @form = Db::EpisodeRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = Db::EpisodeRowsForm.new(episode_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.created")
    end

    def edit
      @episode = Episode.find(params[:id])
      authorize @episode, :edit?
      @work = @episode.work
    end

    def update
      @episode = Episode.find(params[:id])
      authorize @episode, :update?
      @work = @episode.work

      @episode.attributes = episode_params
      @episode.user = current_user

      return render(:edit) unless @episode.valid?
      @episode.save_and_create_activity!

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.updated")
    end

    def hide
      @episode = Episode.find(params[:id])
      authorize @episode, :hide?

      @episode.soft_delete

      flash[:notice] = t("resources.episode.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @episode = Episode.find(params[:id])
      authorize @episode, :destroy?

      @episode.destroy

      flash[:notice] = t("resources.episode.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @episode = Episode.find(params[:id])
      @activities = @episode.db_activities.order(id: :desc)
      @comment = @episode.db_comments.new
    end

    private

    def episode_rows_form_params
      params.require(:db_episode_rows_form).permit(:rows)
    end

    def episode_params
      params.require(:episode).permit(:number, :raw_number, :sort_number, :title, :prev_episode_id)
    end
  end
end
