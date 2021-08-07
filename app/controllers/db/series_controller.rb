# frozen_string_literal: true

module Db
  class SeriesController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @series_list = Series
        .without_deleted
        .order(id: :desc)
        .page(params[:page])
        .per(100)
        .without_count
    end

    def new
      @series = Series.new
      authorize @series
    end

    def create
      @series = Series.new(series_params)
      @series.user = current_user
      authorize @series

      return render(:new, status: :unprocessable_entity) unless @series.valid?

      @series.save_and_create_activity!

      redirect_to db_series_list_path, notice: t("messages._common.created")
    end

    def edit
      @series = Series.without_deleted.find(params[:id])
      authorize @series
    end

    def update
      @series = Series.without_deleted.find(params[:id])
      authorize @series

      @series.attributes = series_params
      @series.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @series.valid?

      @series.save_and_create_activity!

      redirect_to db_edit_series_path(@series), notice: t("messages._common.updated")
    end

    def destroy
      @series = Series.without_deleted.find(params[:id])
      authorize @series

      @series.destroy_in_batches

      redirect_back(
        fallback_location: db_series_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def series_params
      params.require(:series).permit(:name, :name_alter, :name_alter_en, :name_en)
    end
  end
end
