module Db
  class DraftStaffsController < Db::ApplicationController
    permits :person_id, :staff_id, :name, :role, :role_other, :sort_number,
      edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!
    before_action :load_work, only: [:new, :create, :edit, :update]

    def new(staff_id: nil)
      @draft_staff = if staff_id.present?
        @staff = @work.staffs.find(staff_id)
        @work.draft_staffs.new(@staff.attributes.slice(*Staff::DIFF_FIELDS.map(&:to_s)))
      else
        @work.draft_staffs.new
      end
      @draft_staff.build_edit_request
    end

    def create(draft_staff)
      @draft_staff = @work.draft_staffs.new(draft_staff)
      @draft_staff.edit_request.user = current_user
      if @draft_staff.name.blank? && @draft_staff.person.present?
        @draft_staff.name = @draft_staff.person.name
      end

      if draft_staff[:staff_id].present?
        @staff = @work.staffs.find(draft_staff[:staff_id])
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
      @draft_staff = @work.draft_staffs.find(id)
      authorize @draft_staff, :edit?
    end

    def update(id, draft_staff)
      @draft_staff = @work.draft_staffs.find(id)
      authorize @draft_staff, :update?
      if @draft_staff.name.blank? && @draft_staff.person.present?
        @draft_staff.name = @draft_staff.person.name
      end

      if @draft_staff.update(draft_staff)
        flash[:notice] = "編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_staff.edit_request)
      else
        render :edit
      end
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
