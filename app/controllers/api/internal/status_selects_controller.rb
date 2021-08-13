# frozen_string_literal: true

module Api
  module Internal
    class StatusSelectsController < Api::Internal::ApplicationController
      def create
        return head(:unauthorized) unless user_signed_in?

        work = Work.only_kept.find(params[:work_id])
        form = Forms::StatusForm.new(work: work, kind: params[:status_kind])

        if form.valid?
          Updaters::StatusUpdater.new(user: current_user, form: form).call
        end

        render(json: {flash: {type: :notice, message: t("messages._common.updated")}}, status: 201)
      end
    end
  end
end
