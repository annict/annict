# frozen_string_literal: true

module Api::Internal
  class CommentedWorkRecordsController < ApplicationV6Controller
    before_action :authenticate_user!

    def create
      @work = Work.only_kept.find(params[:work_id])
      @form = Forms::WorkRecordForm.new(work: @work, **work_record_form_params)

      if @form.invalid?
        return render json: @form.errors.full_messages, status: :unprocessable_entity
      end

      Creators::WorkRecordCreator.new(user: current_user, form: @form).call

      render(json: {}, status: 201)
    end

    private

    def work_record_form_params
      params.require(:forms_work_record_form).permit(
        :comment, :share_to_twitter,
        :rating_overall, :rating_animation, :rating_character, :rating_story, :rating_music
      )
    end
  end
end
