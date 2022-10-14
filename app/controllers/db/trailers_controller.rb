# frozen_string_literal: true

module Db
  class TrailersController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Work.without_deleted.find(params[:work_id])
      @trailers = @work.trailers.without_deleted.order(:sort_number)
    end

    def new
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::TrailerRowsForm.new
      authorize @form
    end

    def create
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::TrailerRowsForm.new(trailer_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_trailer_list_path(@work), notice: t("messages._common.created")
    end

    def edit
      @trailer = Trailer.without_deleted.find(params[:id])
      authorize @trailer
      @work = @trailer.work
    end

    def update
      @trailer = Trailer.without_deleted.find(params[:id])
      authorize @trailer

      @trailer.attributes = trailer_params
      @trailer.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @trailer.valid?

      @trailer.save_and_create_activity!

      redirect_to db_trailer_list_path(@trailer.work), notice: t("messages._common.updated")
    end

    def destroy
      @trailer = Trailer.without_deleted.find(params[:id])
      authorize @trailer

      @trailer.destroy_in_batches

      redirect_back(
        fallback_location: db_trailer_list_path(@trailer.work),
        notice: t("messages._common.deleted")
      )
    end

    private

    def trailer_rows_form_params
      params.require(:deprecated_db_trailer_rows_form).permit(:rows)
    end

    def trailer_params
      params.require(:trailer).permit(:title, :url, :sort_number)
    end
  end
end
