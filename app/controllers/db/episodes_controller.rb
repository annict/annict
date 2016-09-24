# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    permits :episode_data

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create edit update hide destroy)

    def index(page: nil)
      @episodes = @work.episodes.order(sort_number: :desc).page(page)
    end

    def edit
      @form = DB::EpisodesForm.load(@work)
      # @form = DB::EpisodeForm.new
      # authorize @form, :edit?
    end

    def update(db_multiple_episodes_form)
      @form = DB::MultipleEpisodesForm.load(
        current_user,
        @work,
        db_multiple_episodes_form
      )
      authorize @form, :update?

      saving = @form.valid? &&
        @form.save_and_create_db_activity(current_user, "multiple_episodes.create")
      if saving
        redirect_to db_work_episodes_path(@work), notice: t("resources.episodes.updated")
      else
        render :edit
      end
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
