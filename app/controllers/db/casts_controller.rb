# typed: false
# frozen_string_literal: true

module Db
  class CastsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Work.without_deleted.find(params[:work_id])
      @casts = @work
        .casts
        .without_deleted
        .includes(:person, :character)
        .order(:sort_number)
      @casts_csv = @casts.map { |cast| "#{cast.character.name},#{cast.person.name}" }.join("\n")
    end

    def new
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::CastRowsForm.new
      authorize @form
    end

    def create
      @work = Work.without_deleted.find(params[:work_id])
      @form = Deprecated::Db::CastRowsForm.new(cast_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_cast_list_path(@work), notice: t("resources.cast.created")
    end

    def edit
      @cast = Cast.without_deleted.find(params[:id])
      authorize @cast
      @work = @cast.work
    end

    def update
      @cast = Cast.without_deleted.find(params[:id])
      authorize @cast
      @work = @cast.work

      @cast.attributes = cast_params
      @cast.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @cast.valid?

      @cast.save_and_create_activity!

      redirect_to db_cast_list_path(@work), notice: t("resources.cast.updated")
    end

    def destroy
      @cast = Cast.without_deleted.find(params[:id])
      authorize @cast

      @cast.destroy_in_batches

      redirect_back(
        fallback_location: db_cast_list_path(@cast.work),
        notice: t("messages._common.deleted")
      )
    end

    private

    def cast_rows_form_params
      params.require(:deprecated_db_cast_rows_form).permit(:rows)
    end

    def cast_params
      params.require(:cast).permit(:character_id, :person_id, :name, :name_en, :sort_number)
    end
  end
end
