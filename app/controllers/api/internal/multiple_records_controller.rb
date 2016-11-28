# frozen_string_literal: true

module Api
  module Internal
    class MultipleRecordsController < Api::Internal::ApplicationController
      permits :episode_ids

      before_action :authenticate_user!

      def create(episode_ids)
        records = MultipleRecordsService.new(current_user)
        records.delay.save!(episode_ids)
        keen_client.multiple_records.create(current_user)
        flash[:notice] = t "messages.multiple_records.create.saved"
        head 201
      end
    end
  end
end
