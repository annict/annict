# frozen_string_literal: true

module Db
  class SeriesWorksController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @series = Series.without_deleted.find(params[:series_id])
      @series_works = @series.series_works.preload(:work).without_deleted.sort_season.order(:id)
    end

    def new
      @series = Series.without_deleted.find(params[:series_id])
      @form = Deprecated::Db::SeriesWorkRowsForm.new
      authorize @form, :new?
    end

    def create
      @series = Series.without_deleted.find(params[:series_id])
      @form = Deprecated::Db::SeriesWorkRowsForm.new(series_work_rows_form_params)
      @form.user = current_user
      @form.series = @series
      authorize @form, :create?

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_series_work_list_path(@series), notice: t("messages._common.created")
    end

    def edit
      @series_work = SeriesWork.without_deleted.find(params[:id])

      authorize @series_work

      @series = @series_work.series
    end

    def update
      @series_work = SeriesWork.without_deleted.find(params[:id])

      authorize @series_work

      @series = @series_work.series
      @series_work.attributes = series_work_params
      @series_work.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @series_work.valid?

      @series_work.save_and_create_activity!

      redirect_to db_series_work_list_path(@series), notice: t("messages._common.updated")
    end

    def destroy
      @series_work = SeriesWork.without_deleted.find(params[:id])
      authorize @series_work

      @series_work.destroy_in_batches

      redirect_back(
        fallback_location: db_series_work_list_path(@series_work.series),
        notice: t("messages._common.deleted")
      )
    end

    private

    def series_work_rows_form_params
      params.require(:db_series_work_rows_form).permit(:rows)
    end

    def series_work_params
      params.require(:series_work).permit(:work_id, :summary, :summary_en)
    end
  end
end
