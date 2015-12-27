module Db
  class DraftStaffsController < Db::ApplicationController
    permits :work_id, :staff_id, :name, :role, :role_other,
            edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!
    before_action :load_person, only: [:new, :create, :edit, :update]

    def new(staff_id: nil)
      @draft_staff = if staff_id.present?
        @staff = @person.staffs.find(staff_id)
        @person.draft_staffs.new(@staff.attributes.slice(*Cast::DIFF_FIELDS.map(&:to_s)))
      else
        @person.draft_staffs.new
      end
      @draft_staff.build_edit_request
    end

    def create(draft_staff)
      @draft_staff = @person.draft_staffs.new(draft_staff)
      @draft_staff.edit_request.user = current_user

      if draft_staff[:staff_id].present?
        @staff = @person.staffs.find(draft_staff[:staff_id])
        @draft_staff.origin = @staff
      end

      if @draft_staff.save
        flash[:notice] = "編集リクエストを作成しました"
        redirect_to db_edit_request_path(@draft_staff.edit_request)
      else
        render :new
      end
    end

    def edit(id)
      @draft_staff = @person.draft_staffs.find(id)
      authorize @draft_staff, :edit?
    end

    def update(id, draft_staff)
      @draft_staff = @person.draft_staffs.find(id)
      authorize @draft_staff, :update?

      if @draft_staff.update(draft_staff)
        flash[:notice] = "編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_staff.edit_request)
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
