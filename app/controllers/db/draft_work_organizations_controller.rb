module Db
  class DraftWorkOrganizationsController < Db::ApplicationController
    permits :work_id, :work_organization_id, :role, :role_other,
      edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!
    before_action :load_organization, only: [:new, :create, :edit, :update]

    def new(work_organization_id: nil)
      @draft_wo = if work_organization_id.present?
        @work_organization = @organization.work_organizations.find(work_organization_id)
        fields = WorkOrganization::DIFF_FIELDS.map(&:to_s)
        attrs = @work_organization.attributes.slice(*fields)
        @organization.draft_work_organizations.new(attrs)
      else
        @organization.draft_work_organizations.new
      end
      @draft_wo.build_edit_request
    end

    def create(draft_work_organization)
      @draft_wo = @organization.draft_work_organizations.new(draft_work_organization)
      @draft_wo.edit_request.user = current_user

      if draft_work_organization[:work_organization_id].present?
        work_organization_id = draft_work_organization[:work_organization_id]
        @work_organization = @organization.work_organizations.find(work_organization_id)
        @draft_wo.origin = @work_organization
      end

      if @draft_wo.save
        flash[:notice] = "編集リクエストを作成しました"
        redirect_to db_edit_request_path(@draft_wo.edit_request)
      else
        render :new
      end
    end

    def edit(id)
      @draft_wo = @organization.draft_work_organizations.find(id)
      authorize @draft_wo, :edit?
    end

    def update(id, draft_work_organization)
      @draft_wo = @organization.draft_work_organizations.find(id)
      authorize @draft_wo, :update?

      if @draft_wo.update(draft_work_organization)
        flash[:notice] = "編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_wo.edit_request)
      else
        render :edit
      end
    end

    private

    def load_organization
      @organization = Organization.find(params[:organization_id])
    end
  end
end
