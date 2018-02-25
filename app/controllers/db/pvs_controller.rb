# frozen_string_literal: true

module Db
  class PvsController < Db::ApplicationController
    permits :title, :url, :sort_number

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

    def create(db_pv_rows_form)
      @form = DB::PvRowsForm.new(db_pv_rows_form.permit(:rows).to_h)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!
      @work.purge

      redirect_to db_work_pvs_path(@work), notice: t("messages._common.created")
    end

    def edit
      authorize @pv, :edit?
      @work = @pv.work
    end

    def update(pv)
      authorize @pv, :update?

      @pv.attributes = pv
      @pv.user = current_user

      return render(:edit) unless @pv.valid?
      @pv.save_and_create_activity!
      @pv.work.purge

      redirect_to db_work_pvs_path(@pv.work), notice: t("messages._common.updated")
    end

    def hide
      authorize @pv, :hide?

      @pv.hide!
      @pv.work.purge

      flash[:notice] = t("resources.pv.unpublished")
      redirect_back fallback_location: db_work_pvs_path(@pv.work)
    end

    def destroy
      @pv.destroy
      @pv.work.purge

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
  end
end
