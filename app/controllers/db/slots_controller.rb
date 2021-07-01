# frozen_string_literal: true

module Db
  class SlotsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @work = Anime.without_deleted.find(params[:work_id])
      @slots = @work.slots.without_deleted.eager_load(:channel, :episode, program: :channel)
      @slots = @slots.where(program_id: params[:program_id]) if params[:program_id]
      @slots = @slots.order(started_at: :desc, number: :desc, channel_id: :asc)
      @programs = @work.programs.preload(:channel).without_deleted.where.not(started_at: nil).order(:started_at, :id)
    end

    def new
      @work = Anime.without_deleted.find(params[:work_id])
      @programs = @work.programs.only_kept.where.not(started_at: nil).order(:started_at, :id)
      @form = Db::SlotRowsForm.new
      @form.work = @work
      @form.set_default_rows_by_programs(params[:program_ids]) if params[:program_ids]
      authorize @form
    end

    def create
      @work = Anime.without_deleted.find(params[:work_id])
      @form = Db::SlotRowsForm.new(slot_rows_form)
      @form.user = current_user
      @form.work = @work
      authorize @form

      return render(:new, status: :unprocessable_entity) unless @form.valid?

      ActiveRecord::Base.transaction do
        @form.save!
        @form.reset_number!
      end

      redirect_to db_slot_list_path(@work), notice: t("resources.slot.created")
    end

    def edit
      @slot = Slot.without_deleted.find(params[:id])
      authorize @slot
      @work = @slot.anime
      @programs = @work.programs.order(:started_at)
      @channels = Channel.only_kept.order(:name)
      @episodes = @work.episodes.only_kept.order(sort_number: :desc)
    end

    def update
      @slot = Slot.without_deleted.find(params[:id])
      authorize @slot
      @work = @slot.anime

      @slot.attributes = slot_params
      @slot.user = current_user

      return render(:edit, status: :unprocessable_entity) unless @slot.valid?

      @slot.save_all_and_create_activity!

      redirect_to db_slot_list_path(@work, program_id: @slot.program_id), notice: t("resources.slot.updated")
    end

    def destroy
      @slot = Slot.without_deleted.find(params[:id])
      authorize @slot

      @slot.destroy_in_batches

      redirect_back(
        fallback_location: db_work_list_path,
        notice: t("messages._common.deleted")
      )
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
