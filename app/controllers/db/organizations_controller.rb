# frozen_string_literal: true

module DB
  class OrganizationsController < DB::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update hide destroy)

    def index
      @organizations = Organization.order(id: :desc).page(params[:page])
    end

    def new
      @organization = Organization.new
      authorize @organization, :new?
    end

    def create
      @organization = Organization.new(organization_params)
      @organization.user = current_user
      authorize @organization, :create?

      return render(:new) unless @organization.valid?
      @organization.save_and_create_activity!

      flash[:notice] = t("resources.organization.created")
      redirect_to db_organizations_path
    end

    def edit
      @organization = Organization.find(params[:id])
      authorize @organization, :edit?
    end

    def update
      @organization = Organization.find(params[:id])
      authorize @organization, :update?

      @organization.attributes = organization_params
      @organization.user = current_user

      return render(:edit) unless @organization.valid?
      @organization.save_and_create_activity!

      flash[:notice] = t("resources.organization.updated")
      redirect_to edit_db_organization_path(@organization)
    end

    def hide
      @organization = Organization.find(params[:id])
      authorize @organization, :hide?

      @organization.soft_delete_with_children

      flash[:notice] = t("resources.organization.unpublished")
      redirect_back fallback_location: db_people_path
    end

    def destroy
      @organization = Organization.find(params[:id])
      authorize @organization, :destroy?

      @organization.destroy

      flash[:notice] = t("resources.organization.deleted")
      redirect_back fallback_location: db_people_path
    end

    def activities
      @organization = Organization.find(params[:id])
      @activities = @organization.db_activities.order(id: :desc)
      @comment = @organization.db_comments.new
    end

    private

    def organization_params
      params.require(:organization).permit(
        :name, :name_en, :name_kana, :url, :url_en, :wikipedia_url,
        :wikipedia_url_en, :twitter_username, :twitter_username_en
      )
    end
  end
end
