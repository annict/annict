# frozen_string_literal: true

module Db
  class StaffsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @staffs = @work.staffs.
        includes(:resource).
        order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = Db::StaffRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = Db::StaffRowsForm.new(staff_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.created")
    end

    def edit
      @staff = Staff.find(params[:id])
      authorize @staff, :edit?
      @work = @staff.work
    end

    def update
      @staff = Staff.find(params[:id])
      authorize @staff, :update?
      @work = @staff.work

      @staff.attributes = staff_params
      @staff.user = current_user

      return render(:edit) unless @staff.valid?
      @staff.save_and_create_activity!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.updated")
    end

    def hide
      @staff = Staff.find(params[:id])
      authorize @staff, :hide?

      @staff.hide!

      flash[:notice] = t("resources.staff.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @staff = Staff.find(params[:id])
      authorize @staff, :destroy?

      @staff.destroy

      flash[:notice] = t("resources.staff.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @staff = Staff.find(params[:id])
      @activities = @staff.db_activities.order(id: :desc)
      @comment = @staff.db_comments.new
    end

    private

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
