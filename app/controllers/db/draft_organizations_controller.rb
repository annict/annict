# frozen_string_literal: true

module Db
  class DraftOrganizationsController < Db::ApplicationController
    permits :name, :name_kana, :url, :wikipedia_url, :twitter_username, :organization_id,
      edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!

    def new(organization_id: nil)
      @draft_organization = if organization_id.present?
        @organization = Organization.find(organization_id)
        diff_fields = Organization::DIFF_FIELDS.map(&:to_s)
        DraftOrganization.new(@organization.attributes.slice(*diff_fields))
      else
        DraftOrganization.new
      end
      authorize @draft_organization, :new?
      @draft_organization.build_edit_request
    end

    def create(draft_organization)
      @draft_organization = DraftOrganization.new(draft_organization)
      authorize @draft_organization, :create?
      @draft_organization.edit_request.user = current_user

      if draft_organization[:organization_id].present?
        @organization = Organization.find(draft_organization[:organization_id])
        @draft_organization.origin = @organization
      end

      if @draft_organization.save
        flash[:notice] = "編集リクエストを作成しました"
        redirect_to db_edit_request_path(@draft_organization.edit_request)
      else
        render :new
      end
    end

    def edit(id)
      @draft_organization = DraftOrganization.find(id)
      authorize @draft_organization, :edit?
    end

    def update(id, draft_organization)
      @draft_organization = DraftOrganization.find(id)
      authorize @draft_organization, :update?

      if @draft_organization.update(draft_organization)
        flash[:notice] = "編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_organization.edit_request)
      else
        render :edit
      end
    end
  end
end
