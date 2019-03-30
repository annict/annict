# frozen_string_literal: true

module Db
  class CastsController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_cast, only: %i(edit update destroy hide activities)

    def index
      @casts = @work.casts.
        includes(:person, :character).
        order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @form = DB::CastRowsForm.new
      authorize @form, :new?
    end

    def create
      @form = DB::CastRowsForm.new(cast_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_casts_path(@work), notice: t("resources.cast.created")
    end

    def edit
      authorize @cast, :edit?
      @work = @cast.work
    end

    def update
      authorize @cast, :update?
      @work = @cast.work

      @cast.attributes = cast_params
      @cast.user = current_user

      return render(:edit) unless @cast.valid?
      @cast.save_and_create_activity!

      redirect_to db_work_casts_path(@work), notice: t("resources.cast.updated")
    end

    def hide
      authorize @cast, :hide?

      @cast.hide!

      flash[:notice] = t("resources.cast.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @cast, :destroy?

      @cast.destroy

      flash[:notice] = t("resources.cast.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @cast.db_activities.order(id: :desc)
      @comment = @cast.db_comments.new
    end

    private

    def load_cast
      @cast = Cast.find(params[:id])
    end

    def cast_rows_form_params
      params.require(:db_cast_rows_form).permit(:rows)
    end

    def cast_params
      params.require(:cast).permit(:character_id, :person_id, :name, :name_en, :sort_number)
    end
  end
end
