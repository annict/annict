# frozen_string_literal: true

module Api::Internal
  class RecordsController < ApplicationV6Controller
    include Pundit

    before_action :authenticate_user!

    def create
      @form = Forms::RecordForm.new(record_form_params)

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Creators::RecordCreator.new(user: current_user, form: @form).call

      render(json: {}, status: 201)
    end

    def update
      @record = current_user.records.only_kept.find(params[:record_id])

      authorize @record, :update?

      @form = Forms::RecordForm.new(record_form_params)
      @form.record = @record

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Updaters::RecordUpdater.new(user: current_user, form: @form).call

      render json: {}, status: 200
    end

    private

    def record_form_params
      params.required(:forms_record_form).permit(:body, :episode_id, :instant, :rating, :share_to_twitter, :work_id)
    end
  end
end
