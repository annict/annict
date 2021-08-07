# frozen_string_literal: true

module Db
  class EpisodesController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Anime.without_deleted.find(params[:work_id])
      @episodes = @work.episodes.without_deleted
        .includes(:prev_episode)
        .order(sort_number: :desc)
        .page(params[:page])
        .per(100)
        .without_count
    end

    def new
      @work = Anime.without_deleted.find(params[:work_id])
      @form = Db::EpisodeRowsForm.new
      authorize @form
    end

    def create
      @work = Anime.without_deleted.find(params[:work_id])
      @form = Db::EpisodeRowsForm.new(episode_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_episode_list_path(@work), notice: t("resources.episode.created")
    end

    def edit
      @episode = Episode.without_deleted.find(params[:id])
      authorize @episode
      @work = @episode.anime
    end

    def update
      @episode = Episode.without_deleted.find(params[:id])
      authorize @episode
      @work = @episode.anime

      @episode.attributes = episode_params
      @episode.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @episode.valid?

      @episode.save_and_create_activity!

      redirect_to db_episode_list_path(@work), notice: t("resources.episode.updated")
    end

    def destroy
      @episode = Episode.without_deleted.find(params[:id])
      authorize @episode

      @episode.destroy_in_batches

      redirect_back(
        fallback_location: db_episode_list_path(@episode.anime),
        notice: t("messages._common.deleted")
      )
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
