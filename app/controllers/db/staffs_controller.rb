module Db
  class StaffsController < Db::ApplicationController
    permits :work_id, :name, :role, :role_other

    before_action :authenticate_user!
    before_action :load_person, only: [:index, :new, :create, :edit, :update]

    def index
      @staffs = @person.staffs.order(id: :desc)
    end

    def new
      @staff = @person.staffs.new
      authorize @staff, :new?
    end

    def create(staff)
      @staff = @person.staffs.new(staff)
      authorize @staff, :create?

      if @staff.valid?
        key = "staffs.create"
        @staff.save_and_create_db_activity(current_user, key)
        redirect_to db_person_staffs_path(@person), notice: "登録しました"
      else
        render :new
      end
    end

    def edit(id)
      @staff = @person.staffs.find(id)
      authorize @staff, :edit?
    end

    def update(id, staff)
      @staff = @person.staffs.find(id)
      authorize @staff, :update?
      @staff.attributes = staff

      if @staff.valid?
        key = "staffs.update"
        @staff.save_and_create_db_activity(current_user, key)
        redirect_to db_person_staffs_path(@person), notice: "更新しました"
      else
        render :edit
      end
    end

    private

    def load_person
      @person = Person.find(params[:person_id])
    end
  end
end
