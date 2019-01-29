# frozen_string_literal: true

module Api
  module Internal
    class MultipleRecordsController < Api::Internal::ApplicationController
      permits :episode_ids

      before_action :authenticate_user!

      def create(episode_ids, page_category)
        records = MultipleRecordsService.new(current_user)
        records.save!(episode_ids)
        ga_client.page_category = page_category
        ga_client.events.create(:multiple_records, :create, ds: "internal_api")
        keen_client.publish(:multiple_record_create, via: "internal_api")
        flash[:notice] = t "messages.multiple_records.create.saved"
        head 201
      end
    end
  end
end
