# frozen_string_literal: true

module Db
  class CastsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @casts = @work.casts.
        includes(:person, :character).
        order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = Db::CastRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = Db::CastRowsForm.new(cast_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_casts_path(@work), notice: t("resources.cast.created")
    end

    def edit
      @cast = Cast.find(params[:id])
      authorize @cast, :edit?
      @work = @cast.work
    end

    def update
      @cast = Cast.find(params[:id])
      authorize @cast, :update?
      @work = @cast.work

      @cast.attributes = cast_params
      @cast.user = current_user

      return render(:edit) unless @cast.valid?
      @cast.save_and_create_activity!

      redirect_to db_work_casts_path(@work), notice: t("resources.cast.updated")
    end

    def hide
      @cast = Cast.find(params[:id])
      authorize @cast, :hide?

      @cast.soft_delete

      flash[:notice] = t("resources.cast.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @cast = Cast.find(params[:id])
      authorize @cast, :destroy?

      @cast.destroy

      flash[:notice] = t("resources.cast.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @cast = Cast.find(params[:id])
      @activities = @cast.db_activities.order(id: :desc)
      @comment = @cast.db_comments.new
    end

    private

    def cast_rows_form_params
      params.require(:db_cast_rows_form).permit(:rows)
    end

    def cast_params
      params.require(:cast).permit(:character_id, :person_id, :name, :name_en, :sort_number)
    end
  end
end
