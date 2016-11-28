# frozen_string_literal: true

module Db
  class OrganizationsController < Db::ApplicationController
    permits :name, :name_en, :name_kana, :url, :url_en, :wikipedia_url,
      :wikipedia_url_en, :twitter_username, :twitter_username_en

    before_action :authenticate_user!
    before_action :load_organization, only: %i(edit update hide destroy activities)

    def index(page: nil)
      @organizations = Organization.order(id: :desc).page(page)
    end

    def new
      @organization = Organization.new
      authorize @organization, :new?
    end

    def create(organization)
      @organization = Organization.new(organization)
      @organization.user = current_user
      authorize @organization, :create?

      return render(:new) unless @organization.valid?
      @organization.save_and_create_activity!

      flash[:notice] = t("resources.organization.created")
      redirect_to edit_db_organization_path(@organization)
    end

    def edit
      authorize @organization, :edit?
    end

    def update(organization)
      authorize @organization, :update?

      @organization.attributes = organization
      @organization.user = current_user

      return render(:edit) unless @organization.valid?
      @organization.save_and_create_activity!

      flash[:notice] = t("resources.organization.updated")
      redirect_to edit_db_organization_path(@organization)
    end

    def hide
      authorize @organization, :hide?

      @organization.hide!

      flash[:notice] = t("resources.organization.unpublished")
      redirect_back fallback_location: db_people_path
    end

    def destroy
      authorize @organization, :destroy?

      @organization.destroy

      flash[:notice] = t("resources.organization.deleted")
      redirect_back fallback_location: db_people_path
    end

    def activities
      @activities = @organization.db_activities.order(id: :desc)
      @comment = @organization.db_comments.new
    end

    private

    def load_organization
      @organization = Organization.find(params[:id])
    end
  end
end
