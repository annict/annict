# frozen_string_literal: true

module Db
  class SlotsController < Db::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @slots = @work.slots.eager_load(:channel, :episode, program: :channel)
      @slots = @slots.where(channel_id: params[:channel_id]) if params[:channel_id]
      @slots = @slots.order(number: :desc).order(:channel_id)
    end

    def new
      @work = Work.find(params[:work_id])
      @programs = @work.programs.only_kept.where.not(started_at: nil).order(:started_at, :id)
      @form = Db::SlotRowsForm.new
      @form.work = @work
      @form.set_default_rows_by_programs(params[:program_ids]) if params[:program_ids]
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = Db::SlotRowsForm.new(slot_rows_form)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?

      ActiveRecord::Base.transaction do
        @form.save!
        @form.reset_number!
      end

      redirect_to db_work_slots_path(@work), notice: t("resources.slot.created")
    end

    def edit
      @slot = Slot.find(params[:id])
      authorize @slot, :edit?
      @work = @slot.work
    end

    def update
      @slot = Slot.find(params[:id])
      authorize @slot, :update?
      @work = @slot.work

      @slot.attributes = slot_params
      @slot.user = current_user

      return render(:edit) unless @slot.valid?

      @slot.save_all_and_create_activity!

      redirect_to db_work_slots_path(@work, channel_id: @slot.channel_id), notice: t("resources.slot.updated")
    end

    def hide
      @slot = Slot.find(params[:id])
      authorize @slot, :hide?

      @slot.soft_delete

      flash[:notice] = t("resources.slot.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      @slot = Slot.find(params[:id])
      authorize @slot, :destroy?

      @slot.destroy

      flash[:notice] = t("resources.slot.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @slot = Slot.find(params[:id])
      @activities = @slot.db_activities.order(id: :desc)
      @comment = @slot.db_comments.new
    end

    private

    def slot_params
      params.require(:slot).permit(
        :program_id, :channel_id, :episode_id, :started_at, :number, :rebroadcast,
        :irregular, :time_zone, :shift_time_along_with_after_slots
      )
    end

    def slot_rows_form
      params.require(:db_slot_rows_form).permit(:rows)
    end
  end
end
