# frozen_string_literal: true

module Db
  class StaffsController < Db::ApplicationController
    permits :resource_id, :resource_type, :name, :role, :role_other, :sort_number,
      :name_en, :role_other_en

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

    def create(db_staff_rows_form)
      @form = DB::StaffRowsForm.new(db_staff_rows_form.permit(:rows).to_h)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!
      @work.purge

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.created")
    end

    def edit
      authorize @staff, :edit?
      @work = @staff.work
    end

    def update(staff)
      authorize @staff, :update?
      @work = @staff.work

      @staff.attributes = staff
      @staff.user = current_user

      return render(:edit) unless @staff.valid?
      @staff.save_and_create_activity!
      @work.purge

      redirect_to db_work_staffs_path(@work), notice: t("resources.staff.updated")
    end

    def hide
      authorize @staff, :hide?

      @staff.hide!
      @staff.work.purge

      flash[:notice] = t("resources.staff.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @staff, :destroy?

      @staff.destroy
      @staff.work.purge

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
  end
end
