# frozen_string_literal: true

module DB
  class SeriesWorksController < DB::ApplicationController
    before_action :authenticate_user!

    def index
      @series = Series.find(params[:series_id])
      @series_works = @series.series_works.sort_season
    end

    def new
      @series = Series.find(params[:series_id])
      @form = DB::SeriesWorkRowsForm.new
      authorize @form, :new?
    end

    def create
      @series = Series.find(params[:series_id])
      @form = DB::SeriesWorkRowsForm.new(series_work_rows_form_params)
      @form.user = current_user
      @form.series = @series
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_series_series_works_path(@series)
    end

    def edit
      @series_work = SeriesWork.find(params[:id])
      authorize @series_work, :edit?
      @series = @series_work.series
    end

    def update
      @series_work = SeriesWork.find(params[:id])
      authorize @series_work, :update?
      @series = @series_work.series

      @series_work.attributes = series_work_params
      @series_work.user = current_user

      return render(:edit) unless @series_work.valid?
      @series_work.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_series_series_works_path(@series)
    end

    def hide
      @series_work = SeriesWork.find(params[:id])
      authorize @series_work, :hide?

      @series_work.soft_delete

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_series_series_works_path(@series_work.series)
    end

    def destroy
      @series_work = SeriesWork.find(params[:id])
      authorize @series_work, :destroy?

      @series_work.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_series_series_works_path(@series_work.series)
    end

    def activities
      @series_work = SeriesWork.find(params[:id])
      @activities = @series_work.db_activities.order(id: :desc)
      @comment = @series_work.db_comments.new
    end

    private

    def series_work_rows_form_params
      params.require(:db_series_work_rows_form).permit(:rows)
    end

    def series_work_params
      params.require(:series_work).permit(:series_id, :work_id, :summary, :summary_en)
    end
  end
end
