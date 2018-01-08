# frozen_string_literal: true

module Userland
  class ProjectsController < Userland::ApplicationController
    permits :userland_category_id, :name, :url, :summary, :description, :icon, :available,
      model_name: "UserlandProject"

    before_action :authenticate_user!, only: %i(new create edit update destroy)
    before_action :load_project, only: %i(show edit update destroy)

    def new
      @project = UserlandProject.new
    end

    def create(userland_project)
      @project = UserlandProject.new(userland_project)
      @project.userland_project_members.build(user: current_user)
      @project.detect_locale!(:summary)

      return render(:new) unless @project.valid?

      ActiveRecord::Base.transaction do
        @project.save!(validate: false)
      end

      flash[:notice] = t "messages._common.created"
      redirect_to userland_project_path(@project)
    end

    def edit
      authorize @project, :edit?
    end

    def update(userland_project)
      authorize @project, :update?

      @project.attributes = userland_project
      @project.detect_locale!(:summary)

      if @project.save
        flash[:notice] = t("messages._common.updated")
        redirect_to userland_project_path(@project)
      else
        render :edit
      end
    end

    def destroy
      authorize @project, :destroy?
      @project.destroy
      redirect_to userland_root_path, notice: t("messages._common.deleted")
    end

    private

    def load_project
      @project = UserlandProject.find(params[:id])
    end
  end
end
