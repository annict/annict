# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesController < Api::V1::ApplicationController
        before_action :prepare_params!, only: [:create]

        def create
          work = Work.only_kept.find(@params.work_id)

          form = Forms::StatusForm.new(work: work, kind: @params.kind)

          if form.valid?
            Updaters::StatusUpdater.new(user: current_user, form: form).call
          end

          head 204
        end
      end
    end
  end
end
