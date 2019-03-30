# frozen_string_literal: true

module Db
  class SeriesController < Db::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update hide destroy)

    def index
      @series_list = Series.order(id: :desc).page(params[:page])
    end

    def new
      @series = Series.new
      authorize @series, :new?
    end

    def create
      @series = Series.new(series_params)
      @series.user = current_user
      authorize @series, :create?

      return render(:new) unless @series.valid?
      @series.save_and_create_activity!

      redirect_to db_series_index_path, notice: t("messages._common.created")
    end

    def edit
      @series = Series.find(params[:id])
      authorize @series, :edit?
    end

    def update
      @series = Series.find(params[:id])
      authorize @series, :update?

      @series.attributes = series_params
      @series.user = current_user

      return render(:edit) unless @series.valid?
      @series.save_and_create_activity!

      redirect_to edit_db_series_path(@series), notice: t("messages._common.updated")
    end

    def hide
      @series = Series.find(params[:id])
      authorize @series, :hide?

      @series.hide!

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_series_index_path
    end

    def destroy
      @series = Series.find(params[:id])
      authorize @series, :destroy?

      @series.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_series_index_path
    end

    def activities
      @series = Series.find(params[:id])
      @activities = @series.db_activities.order(id: :desc)
      @comment = @series.db_comments.new
    end

    private

    def series_params
      params.require(:series).permit(:name, :name_en, :name_ro)
    end
  end
end
