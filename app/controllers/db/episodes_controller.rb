# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    permits :number, :raw_number, :sort_number, :title, :prev_episode_id

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_episode, only: %i(edit update hide destroy activities)

    def index(page: nil)
      @episodes = @work.episodes.
        includes(:prev_episode).
        order(sort_number: :desc).
        page(page)
    end

    def new
      @form = DB::EpisodeRowsForm.new
      authorize @form, :new?
    end

    def create(db_episode_rows_form)
      @form = DB::EpisodeRowsForm.new(db_episode_rows_form.permit(:rows).to_h)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.created")
    end

    def edit
      authorize @episode, :edit?
      @work = @episode.work
    end

    def update(episode)
      authorize @episode, :update?
      @work = @episode.work

      @episode.attributes = episode
      @episode.user = current_user

      return render(:edit) unless @episode.valid?
      @episode.save_and_create_activity!

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.updated")
    end

    def hide
      authorize @episode, :hide?

      @episode.hide!

      flash[:notice] = t("resources.episode.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @episode, :destroy?

      @episode.destroy

      flash[:notice] = t("resources.episode.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @episode.db_activities.order(id: :desc)
      @comment = @episode.db_comments.new
    end

    private

    def load_episode
      @episode = Episode.find(params[:id])
    end
  end
end
