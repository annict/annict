# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create edit update hide destroy)

    def index(page: nil)
      @episodes = @work.episodes.order(sort_number: :desc).page(page)
    end

    def new
      @form = DB::EpisodeRowsForm.new
      authorize @form, :new?
    end

    def create(db_episode_rows_form)
      @form = DB::EpisodeRowsForm.new(db_episode_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_episodes_path(@work)
    end

    def edit
      @form = DB::EpisodesForm.load(@work)
      @episodes = @work.episodes.published.order(:sort_number)
      authorize @form, :edit?
    end

    def update(db_episodes_form)
      @form = DB::EpisodesForm.load(@work, db_episodes_form[:episodes])
      authorize @form, :update?

      unless @form.valid?
        @episodes = @work.episodes.published.order(:sort_number)
        return render :edit
      end

      @form.save_and_create_activity!(current_user)
      redirect_to db_work_episodes_path(@work), notice: t("resources.episodes.updated")
    end

    def hide(id)
      @episode = @work.episodes.find(id)
      authorize @episode, :hide?

      @episode.hide!

      redirect_to :back, notice: "エピソードを非公開にしました"
    end

    def destroy(id)
      @episode = @work.episodes.find(id)
      authorize @episode, :destroy?

      @episode.destroy

      redirect_to db_work_episodes_path(@work), notice: "エピソードを削除しました"
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
