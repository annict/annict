# frozen_string_literal: true

module Db
  class StaffsController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_staff, only: %i(edit update destroy hide activities)

    def index
      @staffs = @work.staffs.
        includes(:resource).
        order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @form = DB::StaffRowsForm.new
      authorize @form, :new?
    end

    def create
      @form = DB::StaffRowsForm.new(staff_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.created")
    end

    def edit
      authorize @staff, :edit?
      @work = @staff.work
    end

    def update
      authorize @staff, :update?
      @work = @staff.work

      @staff.attributes = staff_params
      @staff.user = current_user

      return render(:edit) unless @staff.valid?
      @staff.save_and_create_activity!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.updated")
    end

    def hide
      authorize @staff, :hide?

      @staff.hide!

      flash[:notice] = t("resources.staff.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @staff, :destroy?

      @staff.destroy

      flash[:notice] = t("resources.staff.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @staff.db_activities.order(id: :desc)
      @comment = @staff.db_comments.new
    end

    private

    def load_staff
      @staff = Staff.find(params[:id])
    end

    def staff_rows_form_params
      params.require(:db_staff_rows_form).permit(:rows)
    end

    def staff_params
      params.require(:staff).permit(
        :resource_id, :resource_type, :name, :role, :role_other, :sort_number,
        :name_en, :role_other_en
      )
    end
  end
end
