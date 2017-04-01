# frozen_string_literal: true

module Db
  class SeriesWorksController < Db::ApplicationController
    permits :series_id, :work_id, :summary, :summary_en

    before_action :authenticate_user!
    before_action :load_series, only: %i(index new create)
    before_action :load_series_work, only: %i(edit update hide destroy activities)

    def index
      @series_works = @series.series_works.sort_season
    end

    def new
      @form = DB::SeriesWorkRowsForm.new
      authorize @form, :new?
    end

    def create(db_series_work_rows_form)
      @form = DB::SeriesWorkRowsForm.new(db_series_work_rows_form.permit(:rows))
      @form.user = current_user
      @form.series = @series
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_series_series_works_path(@series)
    end

    def edit
      authorize @series_work, :edit?
      @series = @series_work.series
    end

    def update(series_work)
      authorize @series_work, :update?
      @series = @series_work.series

      @series_work.attributes = series_work
      @series_work.user = current_user

      return render(:edit) unless @series_work.valid?
      @series_work.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_series_series_works_path(@series)
    end

    def hide
      authorize @series_work, :hide?

      @series_work.hide!

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_series_series_works_path(@series_work.series)
    end

    def destroy
      authorize @series_work, :destroy?

      @series_work.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_series_series_works_path(@series_work.series)
    end

    def activities
      @activities = @series_work.db_activities.order(id: :desc)
      @comment = @series_work.db_comments.new
    end

    private

    def load_series_work
      @series_work = SeriesWork.find(params[:id])
    end
  end
end
