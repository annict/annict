# frozen_string_literal: true

module Db
  class StaffsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Work.without_deleted.find(params[:work_id])
      @staffs = @work
        .staffs
        .without_deleted
        .includes(:resource)
        .order(:sort_number)
      @staffs_csv = @staffs.map do |staff|
        str = staff.decorate.role_name
        case staff.resource_type
        when "Person"
          str += ",#{staff.resource.name}"
        when "Organization"
          str += ",,#{staff.resource.name}"
        end
        str
      end.join("\n")
    end

    def new
      @work = Work.without_deleted.find(params[:work_id])
      @form = Db::StaffRowsForm.new
      authorize @form
    end

    def create
      @work = Work.without_deleted.find(params[:work_id])
      @form = Db::StaffRowsForm.new(staff_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      @form.save!

      redirect_to db_staff_list_path(@work), notice: t("resources.staff.created")
    end

    def edit
      @staff = Staff.without_deleted.find(params[:id])
      authorize @staff
      @work = @staff.work
    end

    def update
      @staff = Staff.without_deleted.find(params[:id])
      authorize @staff
      @work = @staff.work

      @staff.attributes = staff_params
      @staff.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @staff.valid?

      @staff.save_and_create_activity!

      redirect_to db_staff_list_path(@work), notice: t("resources.staff.updated")
    end

    def destroy
      @staff = Staff.without_deleted.find(params[:id])
      authorize @staff

      @staff.destroy_in_batches

      redirect_back(
        fallback_location: db_staff_list_path(@staff.work),
        notice: t("messages._common.deleted")
      )
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
