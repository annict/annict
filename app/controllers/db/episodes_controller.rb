# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    permits :number, :raw_number, :sort_number, :sc_count, :title,
      :prev_episode_id, :fetch_syobocal

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

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.created")
    end

    def edit(id)
      @episode = @work.episodes.find(id)
      authorize @episode, :edit?
    end

    def update(id, episode)
      @episode = @work.episodes.find(id)
      authorize @episode, :update?

      @episode.attributes = episode
      @episode.user = current_user

      return render(:edit) unless @episode.valid?
      @episode.save_and_create_activity!

      redirect_to db_work_episodes_path(@work), notice: t("resources.episode.updated")
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
