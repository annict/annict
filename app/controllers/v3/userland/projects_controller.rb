# frozen_string_literal: true

module V3::Userland
  class ProjectsController < V3::Userland::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]
    before_action :load_i18n, only: %i[show]

    def new
      @project = UserlandProject.new
    end

    def create
      @project = UserlandProject.new(userland_project_params)
      @project.userland_project_members.build(user: current_user)
      @project.detect_locale!(:summary)

      return render(:new) unless @project.valid?

      ActiveRecord::Base.transaction do
        @project.save!(validate: false)
      end

      flash[:notice] = t "messages._common.created"
      redirect_to userland_project_path(@project)
    end

    def show
      @project = UserlandProject.find(params[:project_id])
    end

    def edit
      @project = UserlandProject.find(params[:project_id])
      authorize @project, :edit?
    end

    def update
      @project = UserlandProject.find(params[:project_id])
      authorize @project, :update?

      @project.attributes = userland_project_params
      @project.detect_locale!(:summary)

      if @project.save
        flash[:notice] = t("messages._common.updated")
        redirect_to userland_project_path(@project)
      else
        render :edit
      end
    end

    def destroy
      @project = UserlandProject.find(params[:project_id])
      authorize @project, :destroy?
      @project.destroy
      redirect_to userland_path, notice: t("messages._common.deleted")
    end

    private

    def load_i18n
      keys = {
        "verb.follow": nil,
        "noun.following": nil
      }

      load_i18n_into_gon keys
    end

    def userland_project_params
      params.require(:userland_project).permit(
        :userland_category_id, :name, :url, :summary, :description, :image, :available
      )
    end
  end
end
