# frozen_string_literal: true

module Db
  class PvsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @pvs = @work.pvs.order(:sort_number)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = DB::PvRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = DB::PvRowsForm.new(pv_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_pvs_path(@work), notice: t("messages._common.created")
    end

    def edit
      @pv = Pv.find(params[:id])
      authorize @pv, :edit?
      @work = @pv.work
    end

    def update
      @pv = Pv.find(params[:id])
      authorize @pv, :update?

      @pv.attributes = pv_params
      @pv.user = current_user

      return render(:edit) unless @pv.valid?
      @pv.save_and_create_activity!

      redirect_to db_work_pvs_path(@pv.work), notice: t("messages._common.updated")
    end

    def hide
      @pv = Pv.find(params[:id])
      authorize @pv, :hide?

      @pv.hide!

      flash[:notice] = t("resources.pv.unpublished")
      redirect_back fallback_location: db_work_pvs_path(@pv.work)
    end

    def destroy
      @pv = Pv.find(params[:id])
      @pv.destroy

      flash[:notice] = t("resources.pv.deleted")
      redirect_back fallback_location: db_work_pvs_path(@pv.work)
    end

    def activities
      @pv = Pv.find(params[:id])
      @activities = @pv.db_activities.order(id: :desc)
      @comment = @pv.db_comments.new
    end

    private

    def pv_rows_form_params
      params.require(:db_pv_rows_form).permit(:rows)
    end

    def pv_params
      params.require(:pv).permit(:title, :url, :sort_number)
    end
  end
end
