# frozen_string_literal: true

module Db
  class OrganizationsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update destroy)

    def index
      @organizations = Organization.
        without_deleted.
        order(id: :desc).
        page(params[:page]).
        per(100)
    end

    def new
      @form = Db::OrganizationRowsForm.new
      authorize_db_resource @form
    end

    def create
      @form = Db::OrganizationRowsForm.new(organization_rows_form_params)
      @form.user = current_user
      authorize_db_resource @form

      return render(:new) unless @form.valid?

      @form.save!

      redirect_to db_organization_list_path, notice: t("resources.person.created")
    end

    def edit
      @organization = Organization.without_deleted.find(params[:id])
      authorize_db_resource @organization
    end

    def update
      @organization = Organization.without_deleted.find(params[:id])
      authorize_db_resource @organization

      @organization.attributes = organization_params
      @organization.user = current_user

      return render(:edit) unless @organization.valid?

      @organization.save_and_create_activity!

      redirect_to db_edit_organization_path(@organization), notice: t("resources.person.updated")
    end

    def destroy
      @organization = Organization.without_deleted.find(params[:id])
      authorize_db_resource @organization

      @organization.soft_delete

      redirect_back(
        fallback_location: db_organization_list_path,
        notice: t("resources.cast.deleted")
      )
    end

    private

    def organization_rows_form_params
      params.require(:db_organization_rows_form).permit(:rows)
    end

    def organization_params
      params.require(:organization).permit(
        :name, :name_en, :name_kana, :url, :url_en, :wikipedia_url,
        :wikipedia_url_en, :twitter_username, :twitter_username_en
      )
    end
  end
end
