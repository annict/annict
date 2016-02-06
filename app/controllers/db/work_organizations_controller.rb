module Db
  class WorkOrganizationsController < Db::ApplicationController
    permits :organization_id, :name, :role, :role_other, :sort_number

    before_action :authenticate_user!
    before_action :load_work, only: [:index, :new, :create, :edit, :update, :destroy]

    def index
      @work_organizations = @work.work_organizations.order(:sort_number)
    end

    def new
      @work_organization = @work.work_organizations.new
      authorize @work_organization, :new?
    end

    def create(work_organization)
      @work_organization = @work.work_organizations.new(work_organization)
      authorize @work_organization, :create?

      if @work_organization.valid?
        key = "work_organizations.create"
        @work_organization.save_and_create_db_activity(current_user, key)
        path = db_work_work_organizations_path(@work)
        redirect_to path, notice: "登録しました"
      else
        render :new
      end
    end

    def edit(id)
      @work_organization = @work.work_organizations.find(id)
      authorize @work_organization, :edit?
    end

    def update(id, work_organization)
      @work_organization = @work.work_organizations.find(id)
      authorize @work_organization, :update?
      @work_organization.attributes = work_organization

      if @work_organization.valid?
        key = "work_organizations.update"
        @work_organization.save_and_create_db_activity(current_user, key)
        path = db_work_work_organizations_path(@work)
        redirect_to path, notice: "更新しました"
      else
        render :edit
      end
    end

    def destroy(id)
      @work_organization = @work.work_organizations.find(id)
      authorize @work_organization, :destroy?

      @work_organization.destroy

      redirect_to :back, notice: "削除しました"
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
