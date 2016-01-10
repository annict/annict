module Db
  class OrganizationsController < Db::ApplicationController
    permits :name, :url, :wikipedia_url, :twitter_username

    before_action :authenticate_user!, only: [:new, :create, :edit, :update, :destroy]

    def index(page: nil)
      @organizations = Organization.order(id: :desc).page(page)
    end

    def new
      @organization = Organization.new
      authorize @organization, :new?
    end

    def create(organization)
      @organization = Organization.new(organization)
      authorize @organization, :create?

      if @organization.save_and_create_db_activity(current_user, "organizations.create")
        redirect_to edit_db_organization_path(@organization), notice: "登録しました"
      else
        render :new
      end
    end

    def edit(id)
      @organization = Organization.find(id)
      authorize @organization, :edit?
    end

    def update(id, organization)
      @organization = Organization.find(id)
      authorize @organization, :update?

      @organization.attributes = organization
      if @organization.save_and_create_db_activity(current_user, "organizations.update")
        redirect_to edit_db_organization_path(@organization), notice: "更新しました"
      else
        render :edit
      end
    end
  end
end
