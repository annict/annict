# frozen_string_literal: true

module Db
  class StaffsController < Db::ApplicationController
    permits :resource_id, :resource_type, :name, :role, :role_other, :sort_number

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create edit update hide destroy)

    def index
      @staffs = @work.staffs.
        includes(:resource).
        order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @form = DB::StaffRowsForm.new
      authorize @form, :new?
    end

    def create(db_staff_rows_form)
      @form = DB::StaffRowsForm.new(db_staff_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.created")
    end

    def edit(id)
      @staff = @work.staffs.find(id)
      authorize @staff, :edit?
    end

    def update(id, staff)
      @staff = @work.staffs.find(id)
      authorize @staff, :update?

      @staff.attributes = staff
      @staff.user = current_user

      return render(:edit) unless @staff.valid?
      @staff.save_and_create_activity!

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.updated")
    end

    def hide(id)
      @staff = @work.staffs.find(id)
      authorize @staff, :hide?

      @staff.hide!

      flash[:notice] = t("resources.staff.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy(id)
      @staff = @work.staffs.find(id)
      authorize @staff, :destroy?

      @staff.destroy

      flash[:notice] = t("resources.staff.deleted")
      redirect_back fallback_location: db_works_path
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
