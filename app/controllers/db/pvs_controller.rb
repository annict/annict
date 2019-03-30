# frozen_string_literal: true

module Db
  class PvsController < Db::ApplicationController
    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_pv, only: %i(edit update hide destroy activities)

    def index
      @pvs = @work.pvs.order(:sort_number)
    end

    def new
      @form = DB::PvRowsForm.new
      authorize @form, :new?
    end

    def create
      @form = DB::PvRowsForm.new(pv_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_pvs_path(@work), notice: t("messages._common.created")
    end

    def edit
      authorize @pv, :edit?
      @work = @pv.work
    end

    def update
      authorize @pv, :update?

      @pv.attributes = pv_params
      @pv.user = current_user

      return render(:edit) unless @pv.valid?
      @pv.save_and_create_activity!

      redirect_to db_work_pvs_path(@pv.work), notice: t("messages._common.updated")
    end

    def hide
      authorize @pv, :hide?

      @pv.hide!

      flash[:notice] = t("resources.pv.unpublished")
      redirect_back fallback_location: db_work_pvs_path(@pv.work)
    end

    def destroy
      @pv.destroy

      flash[:notice] = t("resources.pv.deleted")
      redirect_back fallback_location: db_work_pvs_path(@pv.work)
    end

    def activities
      @activities = @pv.db_activities.order(id: :desc)
      @comment = @pv.db_comments.new
    end

    private

    def load_pv
      @pv = Pv.find(params[:id])
    end

    def pv_rows_form_params
      params.require(:db_pv_rows_form).permit(:rows)
    end

    def pv_params
      params.require(:pv).permit(:title, :url, :sort_number)
    end
  end
end
